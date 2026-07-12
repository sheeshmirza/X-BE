package service

import (
	"X-BE/internal/repository"
	"X-BE/internal/zoho"
)

type Services struct {
	Career     *CareerService
	Newsletter *NewsletterService
	Sale       *SaleService
	Zoho       *ZohoService
}

func NewServices(repos *repository.Repositories) *Services {
	if repos == nil {
		return &Services{}
	}
	newsletterService := NewNewsletterService(repos.Newsletter)
	newsletterService.zoho = repos.Zoho
	careerService := NewCareerService(repos.Career, newsletterService)
	careerService.zoho = repos.Zoho
	saleService := NewSaleService(repos.Sale, newsletterService)
	saleService.zoho = repos.Zoho
	return &Services{
		Career:     careerService,
		Newsletter: newsletterService,
		Sale:       saleService,
		Zoho:       NewZohoService(repos.Zoho),
	}
}

type CareerService struct {
	repo       *repository.CareerRepository
	newsletter *NewsletterService
	zoho       *zoho.Client
}

type NewsletterService struct {
	repo *repository.NewsletterRepository
	zoho *zoho.Client
}

type SaleService struct {
	repo       *repository.SaleRepository
	newsletter *NewsletterService
	zoho       *zoho.Client
}

type ZohoService struct {
	client *zoho.Client
}
