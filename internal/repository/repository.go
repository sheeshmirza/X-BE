package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"X-BE/internal/config"
	"X-BE/internal/models"
	"X-BE/internal/zoho"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Repositories struct {
	Career     *CareerRepository
	Newsletter *NewsletterRepository
	Sale       *SaleRepository
	Zoho       *zoho.Client
}

type PersistenceOutcome struct {
	ZohoStatus string
	Errors     []string
}

func NewRepositories(db *mongo.Database, cfg config.Config) *Repositories {
	zohoClient := zoho.NewClient(cfg, db)
	return &Repositories{
		Career: &CareerRepository{
			collection: db.Collection("career_applications"),
			zoho:       zohoClient,
			sheet:      cfg.ZohoSheets.Career,
		},
		Newsletter: &NewsletterRepository{
			collection: db.Collection("newsletter_subscribers"),
			zoho:       zohoClient,
			sheet:      cfg.ZohoSheets.Newsletter,
		},
		Sale: &SaleRepository{
			collection: db.Collection("sales_leads"),
			zoho:       zohoClient,
			sheet:      cfg.ZohoSheets.Sale,
		},
		Zoho: zohoClient,
	}
}

type CareerRepository struct {
	collection *mongo.Collection
	zoho       *zoho.Client
	sheet      config.ZohoSheetConfig
}

func (r *CareerRepository) Create(ctx context.Context, app models.CareerApplication) (PersistenceOutcome, error) {
	_, err := r.collection.InsertOne(ctx, app)
	if err != nil {
		return PersistenceOutcome{}, err
	}
	row := map[string]string{
		"Date":       app.Date.Format(time.RFC3339),
		"Name":       app.Name,
		"Email":      app.Email,
		"Phone":      app.Phone,
		"City":       app.City,
		"State":      app.State,
		"Country":    app.Country,
		"Department": app.Department,
		"Message":    app.Message,
	}
	return appendToZoho(ctx, r.zoho, r.sheet, row)
}

type NewsletterRepository struct {
	collection *mongo.Collection
	zoho       *zoho.Client
	sheet      config.ZohoSheetConfig
}

func (r *NewsletterRepository) Subscribe(ctx context.Context, email string) (PersistenceOutcome, error) {
	now := time.Now().UTC()
	update := bson.M{
		"$set": bson.M{
			"email":    email,
			"isActive": true,
			"source":   "website",
		},
		"$setOnInsert": bson.M{
			"date": now,
		},
	}
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"email": email},
		update,
		options.Update().SetUpsert(true),
	)
	if err != nil {
		return PersistenceOutcome{}, err
	}
	row := map[string]string{
		"Date":  now.Format(time.RFC3339),
		"Email": email,
	}
	return appendToZoho(ctx, r.zoho, r.sheet, row)
}

func (r *NewsletterRepository) IsSubscribed(ctx context.Context, email string) (bool, error) {
	count, err := r.collection.CountDocuments(ctx, bson.M{"email": email, "isActive": true})
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *NewsletterRepository) Unsubscribe(ctx context.Context, email string) (bool, error) {
	result, err := r.collection.UpdateOne(
		ctx,
		bson.M{"email": email, "isActive": true},
		bson.M{"$set": bson.M{"isActive": false}},
	)
	if err != nil {
		return false, err
	}
	return result.ModifiedCount > 0, nil
}

type SaleRepository struct {
	collection *mongo.Collection
	zoho       *zoho.Client
	sheet      config.ZohoSheetConfig
}

func (r *SaleRepository) Create(ctx context.Context, lead models.SaleLead) (PersistenceOutcome, error) {
	_, err := r.collection.InsertOne(ctx, lead)
	if err != nil {
		return PersistenceOutcome{}, err
	}
	row := map[string]string{
		"Date":     lead.Date.Format(time.RFC3339),
		"Name":     lead.Name,
		"Email":    lead.Email,
		"Query":    lead.Query,
		"Phone":    "",
		"Budget":   "",
		"Timeline": "",
	}
	return appendToZoho(ctx, r.zoho, r.sheet, row)
}

func appendToZoho(ctx context.Context, client *zoho.Client, sheet config.ZohoSheetConfig, row map[string]string) (PersistenceOutcome, error) {
	if client == nil || !client.Enabled() {
		return PersistenceOutcome{ZohoStatus: "skipped"}, nil
	}
	err := client.AppendRow(ctx, sheet, row)
	if err == nil {
		return PersistenceOutcome{ZohoStatus: "saved"}, nil
	}
	if errors.Is(err, zoho.ErrNotConfigured) || errors.Is(err, zoho.ErrSheetNotConfigured) || errors.Is(err, zoho.ErrTokenNotFound) {
		return PersistenceOutcome{ZohoStatus: "skipped"}, nil
	}
	return PersistenceOutcome{
		ZohoStatus: "failed",
		Errors:     []string{fmt.Sprintf("Zoho: %v", err)},
	}, nil
}
