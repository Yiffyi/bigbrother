package ctrl

import (
	"context"
	"errors"
	"time"

	"github.com/yiffyi/bigbrother/ppp/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type UserController struct {
	db *gorm.DB
}

func NewUserController(db *gorm.DB) *UserController {
	c := UserController{
		db,
	}

	return &c
}

func (c *UserController) GetUserBySubscriptionToken(token string) (*model.User, error) {
	ctx := context.Background()
	user, err := gorm.G[model.User](c.db).Where("subs_token = ?", token).First(ctx)
	if err != nil {
		return nil, err
	}

	return &user, err
}

func (c *UserController) GetUserByID(id string) (*model.User, error) {
	ctx := context.Background()
	user, err := gorm.G[model.User](c.db).Where("id = ?", id).First(ctx)
	if err != nil {
		return nil, err
	}

	return &user, err
}

func (c *UserController) CreateInvoice(userId uint, amount int, days int) (*model.Invoice, error) {
	ctx := context.Background()

	invoice := model.Invoice{
		Amount: amount, Days: days, State: model.INVOICE_STATE_INVALID,
		Claimed: false,
		UserID:  userId,
	}

	err := gorm.G[model.Invoice](c.db).Create(ctx, &invoice) // pass pointer of data to Create

	return &invoice, err
}

func (c *UserController) UpdateInvoiceState(invoiceId uint, newState model.InvoiceState) error {
	ctx := context.Background()

	cnt, err := gorm.G[model.Invoice](c.db).
		Where("id = ?", invoiceId). // maybe check if it is claimed
		Update(ctx, "state", newState)
	if err != nil {
		return err
	}

	if cnt != 1 {
		return errors.New("could not update target invoice")
	}

	return nil
}

func (c *UserController) ExtendSubscription(invoiceId uint) (user *model.User, err error) {
	ctx := context.Background()

	// Basic transaction
	err = c.db.Transaction(func(tx *gorm.DB) error {
		claimedCnt, err := gorm.G[model.Invoice](tx).
			Where(
				"id = ? AND state IN ? AND NOT claimed",
				invoiceId,
				[]model.InvoiceState{model.INVOICE_STATE_SUCCESS, model.INVOICE_STATE_FINISH},
			).
			Update(ctx, "claimed", true)

		if err != nil {
			return err
		}

		if claimedCnt != 1 {
			return errors.New("could claim target invoice to extend subscription")
		}

		invoice, err := gorm.G[model.Invoice](tx).
			Joins(clause.JoinTarget{Association: "User"}, nil).
			Where(
				"id = ? AND state IN ? AND NOT claimed",
				invoiceId,
				[]model.InvoiceState{model.INVOICE_STATE_SUCCESS, model.INVOICE_STATE_FINISH},
			).
			Take(ctx)

		if err != nil {
			return err
		}

		user = &invoice.User

		now := time.Now()
		if invoice.User.SubsEndAt.Before(now) { // no valid subscription for now
			cnt, err := gorm.G[model.User](tx).
				Where("id = ?", invoice.UserID).
				Updates(ctx, model.User{
					SubsBeginAt: now,
					SubsEndAt:   now.Add(time.Duration(invoice.Days) * time.Hour * 24),
				})

			if err != nil {
				return err
			}

			if cnt != 1 {
				return errors.New("could not update user")
			}

		} else { // we have an on-going subscription
			cnt, err := gorm.G[model.User](tx).
				Where("id = ?", invoice.UserID).
				Update(ctx, "subs_end_at", gorm.Expr("subs_end_at + ?", time.Duration(invoice.Days)*time.Hour*24))

			if err != nil {
				return err
			}

			if cnt != 1 {
				return errors.New("could not update user")
			}
		}

		// return nil will commit the whole transaction
		return nil
	})

	return
}
