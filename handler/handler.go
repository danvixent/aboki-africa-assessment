package handler

import (
	"context"
	"crypto/rand"
	"io"

	app "github.com/danvixent/aboki-africa-assessment"
	"github.com/danvixent/aboki-africa-assessment/datastore/postgres"
	"github.com/danvixent/aboki-africa-assessment/errors"
	"github.com/jackc/pgx/v4"
	log "github.com/sirupsen/logrus"
)

type Handler struct {
	userRepository         app.UserRepository
	userReferralRepository app.UserReferralRepository
	userPointRepository    app.UserPointRepository
	beginTxFunc            func() (pgx.Tx, error)
}

func NewHandler(userRepository app.UserRepository, userReferralRepository app.UserReferralRepository, userPointRepository app.UserPointRepository, beginTxFunc func() (pgx.Tx, error)) *Handler {
	return &Handler{
		userRepository:         userRepository,
		userReferralRepository: userReferralRepository,
		userPointRepository:    userPointRepository,
		beginTxFunc:            beginTxFunc,
	}
}

func (h *Handler) RegisterUser(ctx context.Context, input *UserRequest, logger *log.Entry) (*app.User, error) {
	tx, err := h.beginTxFunc()
	if err != nil {
		logger.WithError(err).Error("failed to start transaction")
		return nil, errors.ErrGeneric
	}
	defer tx.Rollback(ctx)

	ctx = context.WithValue(ctx, app.TxContextKey, tx)

	user := &app.User{
		Name:         input.Name,
		Email:        input.Email,
		ReferralCode: GenReferralCode(6),
	}

	err = h.userRepository.CreateUser(ctx, user)
	if err != nil {
		logger.WithError(err).Error("failed to create user")
		return nil, errors.ErrCreateUserFailed
	}

	userPoint := &app.UserPoints{
		UserID: user.ID,
		Points: 0,
	}

	err = h.userPointRepository.CreateUserPoint(ctx, userPoint)
	if err != nil {
		logger.WithError(err).Error("failed to create user point balance")
		return nil, errors.ErrGeneric
	}

	if input.ReferralCode != nil {
		referrer, err := h.userRepository.FindUserByReferralCode(ctx, *input.ReferralCode)
		if err != nil {
			logger.WithError(err).Error("failed to find user by referral code")
			return nil, errors.ErrGeneric
		}

		userReferral := &app.UserReferral{
			ReferrerID: referrer.ID,
			RefereeID:  user.ID,
			PaidOut:    false,
		}

		err = h.userReferralRepository.CreateUserReferral(ctx, userReferral)
		if err != nil {
			logger.WithError(err).Error("failed to save user referral")
			return nil, errors.ErrGeneric
		}

		unpaidCount, err := h.userReferralRepository.GetUnpaidUserReferralCount(ctx, referrer.ID)
		if err != nil {
			logger.WithError(err).Error("failed to get unpaid user referral count")
			return nil, errors.ErrGeneric
		}

		if unpaidCount == 3 {
			err = h.userPointRepository.CreditUser(ctx, referrer.ID, 50)
			if err != nil {
				logger.WithError(err).Error("failed credit user referrer")
				return nil, errors.ErrGeneric
			}

			err = h.userReferralRepository.MarkPendingReferralsAsPaid(ctx, referrer.ID)
			if err != nil {
				logger.WithError(err).Error("failed to mark pending referrals as paid")
				return nil, errors.ErrGeneric
			}
		}
	}

	if err = tx.Commit(ctx); err != nil {
		logger.WithError(err).Error("failed to commit transaction")
		return nil, errors.ErrGeneric
	}

	return user, nil
}

// GenReferralCode helps to generate reference code.
func GenReferralCode(max int) string {
	table := []byte("abcdefghijklmnopqrstuvwxyz-ABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")
	b := make([]byte, max)
	n, err := io.ReadAtLeast(rand.Reader, b, max)
	if n != max {
		panic(err)
	}
	for i := 0; i < len(b); i++ {
		b[i] = table[int(b[i])%len(table)]
	}
	return string(b)
}

func (h *Handler) TransferPoints(ctx context.Context, input *TransferPointsRequest, logger *log.Entry) error {
	tx, err := h.beginTxFunc()
	if err != nil {
		logger.WithError(err).Error("failed to start transaction")
		return errors.ErrGeneric
	}
	defer tx.Rollback(ctx)

	ctx = context.WithValue(ctx, app.TxContextKey, tx)

	balance, err := h.userPointRepository.GetUserPointsBalance(ctx, input.UserID)
	if err != nil {
		logger.WithError(err).Error("failed to get user points balance")
		return errors.ErrGeneric
	}

	if balance < input.Points {
		return errors.ErrInsufficientFunds
	}

	// we get the total before recording the transaction so we can determine if the total transferred points
	// was previously less than 200
	totalTransferredPoints, err := h.userPointRepository.GetUserTotalTransferredPoints(ctx, input.UserID)
	if err != nil {
		logger.WithError(err).Error("failed to get user total transferred points")
		return errors.ErrGeneric
	}

	err = h.userPointRepository.TransferPoints(ctx, input.UserID, input.RecipientUserID, input.Points)
	if err != nil {
		return errors.Wrap(err, "transfer points failed")
	}

	// record the point transfer
	txn := &app.Transaction{
		UserID:          input.UserID,
		RecipientUserID: input.RecipientUserID,
		Points:          input.Points,
	}

	if err = h.userPointRepository.CreatePointTransaction(ctx, txn); err != nil {
		logger.WithError(err).Error("failed to create transaction")
		return errors.ErrCreditUserFailed
	}

	// if it was previously less than 200 and now it's greater than 200, credit the referrer who
	// referred this user with 50 points.
	if totalTransferredPoints <= 200 && totalTransferredPoints+input.Points > 200 {
		referrer, err := h.userReferralRepository.GetUserReferrer(ctx, input.UserID)
		if err != nil {
			// if it's pgx.ErrNoRows, it means this user wasn't referred by anyone
			if errors.Is(err, pgx.ErrNoRows) {
				return nil
			}
			logger.WithError(err).Error("failed to find user referrer")
			return errors.ErrGeneric
		}

		bonus := &app.ReferredUserTransactionBonus{
			ReferrerID: referrer.ID,
			RefereeID:  input.UserID,
			PaidOut:    false,
		}

		err = h.userReferralRepository.CreateReferredUserTransactionBonus(ctx, bonus)
		if err != nil && !postgres.IsDuplicateError(err) {
			logger.WithError(err).Error("failed to create referred user transaction bonus")
			return errors.ErrGeneric
		}

		bonuses, err := h.userReferralRepository.GetUnpaidReferredUserTransactionBonus(ctx, referrer.ID)
		if err != nil {
			logger.WithError(err).Error("failed to get unpaid referred user transaction bonuses")
			return errors.ErrGeneric
		}

		if len(bonuses) == 3 {
			bonusIDs := make([]string, len(bonuses))
			for i, b := range bonuses {
				bonusIDs[i] = b.ID
			}

			err = h.userReferralRepository.PayReferralsTransactionsBonuses(ctx, bonusIDs)
			if err != nil {
				logger.WithError(err).Error("failed to pay referrals transactions bonuses")
				return errors.ErrGeneric
			}

			err = h.userPointRepository.CreditUser(ctx, referrer.ID, 50)
			if err != nil {
				logger.WithError(err).Error("failed to credit referrer with referred user transaction bonuses")
				return errors.ErrCreditUserFailed
			}
		}
	}

	if err = tx.Commit(ctx); err != nil {
		logger.WithError(err).Error("failed to commit transaction")
		return errors.ErrGeneric
	}

	return nil
}
