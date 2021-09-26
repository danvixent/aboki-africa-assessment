package handler

import (
	"context"
	"crypto/rand"
	app "github.com/danvixent/aboki-africa-assessment"
	"github.com/danvixent/aboki-africa-assessment/datastore/postgres"
	"github.com/danvixent/aboki-africa-assessment/errors"
	"github.com/jackc/pgx"
	log "github.com/sirupsen/logrus"
	"io"
)

type Handler struct {
	client *postgres.Client
}

func (h *Handler) RegisterUser(ctx context.Context, input *UserRequest) (*app.User, error) {
	logger := log.WithFields(map[string]interface{}{})
	tx, err := h.client.BeginTx()
	if err != nil {
		logger.WithError(err).Error("failed to start transaction")
		return nil, errors.ErrGeneric
	}
	defer tx.Rollback(ctx)

	user := &app.User{
		Name:         input.Name,
		Email:        input.Email,
		ReferralCode: GenReferralCode(6),
	}

	userResource := postgres.NewUserResource(tx)
	err = userResource.CreateUser(ctx, user)
	if err != nil {
		logger.WithError(err).Error("failed to create user")
		return nil, errors.ErrCreateUserFailed
	}

	userReferralResource := postgres.NewUserReferralResource(tx)
	if input.ReferralCode != nil {
		referrer, err := userResource.FindUserByReferralCode(ctx, *input.ReferralCode)
		if err != nil {
			logger.WithError(err).Error("failed to find user by referral code")
			return nil, errors.ErrGeneric
		}

		userReferral := &app.UserReferral{
			ReferrerID: referrer.ID,
			RefereeID:  user.ID,
			PaidOut:    false,
		}

		err = userReferralResource.CreateUserReferral(ctx, userReferral)
		if err != nil {
			logger.WithError(err).Error("failed to save user referral")
			return nil, errors.ErrGeneric
		}

		unpaidCount, err := userReferralResource.GetUnpaidUserReferralCount(ctx, referrer.ID)
		if err != nil {
			logger.WithError(err).Error("failed to get unpaid user referral count")
			return nil, errors.ErrGeneric
		}

		if unpaidCount == 3 {
			err = postgres.NewUserPointsResource(tx).CreditUser(ctx, referrer.ID, 50)
			if err != nil {
				logger.WithError(err).Error("failed credit user referrer")
				return nil, errors.ErrGeneric
			}

			err = userReferralResource.MarkPendingReferralsAsPaid(ctx, referrer.ID)
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

func (h *Handler) TransferPoints(ctx context.Context, input *TransferPointsRequest) error {
	logger := log.WithFields(map[string]interface{}{})
	tx, err := h.client.BeginTx()
	if err != nil {
		logger.WithError(err).Error("failed to start transaction")
		return errors.ErrGeneric
	}
	defer tx.Rollback(ctx)

	//cctx := context.WithValue(ctx, "pgxTX", tx)

	userPointResource := postgres.NewUserPointsResource(tx)
	balance, err := userPointResource.GetUserPointsBalance(ctx, input.UserID)
	if err != nil {
		logger.WithError(err).Error("failed to get user points balance")
		return errors.ErrGeneric
	}

	if balance < input.Points {
		return errors.ErrInsufficientFunds
	}

	totalTransferredPoints, err := userPointResource.GetUserTotalTransferredPoints(ctx, input.UserID)
	if err != nil {
		logger.WithError(err).Error("failed to get user total transferred points")
		return errors.ErrGeneric
	}

	if err = userPointResource.DebitUser(ctx, input.Points, input.UserID); err != nil {
		logger.WithError(err).Error("failed to debit sender")
		return errors.ErrDebitUserFailed
	}

	if err = userPointResource.CreditUser(ctx, input.RecipientUserID, input.Points); err != nil {
		logger.WithError(err).Error("failed to credit recipient")
		return errors.ErrCreditUserFailed
	}

	txn := &app.Transaction{
		UserID:          input.UserID,
		RecipientUserID: input.RecipientUserID,
		Points:          input.Points,
	}

	if err = userPointResource.CreatePointTransaction(ctx, txn); err != nil {
		logger.WithError(err).Error("failed to create transaction")
		return errors.ErrCreditUserFailed
	}

	if totalTransferredPoints < 200 && totalTransferredPoints+input.Points > 200 {
		userReferralResource := postgres.NewUserReferralResource(tx)
		referrer, err := userReferralResource.GetUserReferrer(ctx, input.UserID)
		if err != nil {
			// if it's pgx.ErrNoRows, it means this user wasn't referred by anyone
			if !errors.Is(err, pgx.ErrNoRows) {
				logger.WithError(err).Error("failed to find user referrer")
				return errors.ErrGeneric
			}
		}

	}

	return nil
}
