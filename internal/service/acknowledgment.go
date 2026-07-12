package service

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"X-BE/internal/zoho"
)

func (s *CareerService) sendAcknowledgmentEmail(email, name, department string, appliedAt time.Time) {
	if s == nil || s.zoho == nil {
		return
	}
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
		defer cancel()
		subject := fmt.Sprintf("Your Application for %s Position - XEFORT SOLUTIONS", strings.TrimSpace(department))
		content := fmt.Sprintf("Hi %s,\n\nThank you for applying for the %s position at XEFORT SOLUTIONS. We received your application on %s and our team will review it soon.\n\nRegards,\nXEFORT SOLUTIONS Team", strings.TrimSpace(name), strings.TrimSpace(department), appliedAt.Format(time.RFC1123))
		err := s.zoho.SendEmail(ctx, zoho.EmailRequest{
			FromAddress: "careers@xefort.com",
			To:          []string{email},
			Subject:     subject,
			Content:     content,
			ContentType: "text",
		})
		if err != nil {
			log.Printf("[Career] acknowledgment email failed: %v", err)
		}
	}()
}

func (s *NewsletterService) sendAcknowledgmentEmail(email string) {
	if s == nil || s.zoho == nil {
		return
	}
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
		defer cancel()
		err := s.zoho.SendEmail(ctx, zoho.EmailRequest{
			FromAddress: "marketing@xefort.com",
			To:          []string{email},
			Subject:     "Welcome to XEFORT SOLUTIONS Newsletter",
			Content:     fmt.Sprintf("Thank you for subscribing to XEFORT SOLUTIONS newsletter. Updates will be sent to %s.", email),
			ContentType: "text",
		})
		if err != nil {
			log.Printf("[Newsletter] acknowledgment email failed: %v", err)
		}
	}()
}

func (s *SaleService) sendAcknowledgmentEmail(email, name, query string) {
	if s == nil || s.zoho == nil {
		return
	}
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
		defer cancel()
		err := s.zoho.SendEmail(ctx, zoho.EmailRequest{
			FromAddress: "info@xefort.com",
			To:          []string{email},
			Subject:     fmt.Sprintf("XEFORT SOLUTIONS - Thank you for your inquiry, %s", strings.TrimSpace(name)),
			Content:     fmt.Sprintf("Hi %s,\n\nThank you for contacting XEFORT SOLUTIONS. We received your inquiry: \"%s\". Our team will respond shortly.\n\nRegards,\nXEFORT SOLUTIONS Team", strings.TrimSpace(name), strings.TrimSpace(query)),
			ContentType: "text",
		})
		if err != nil {
			log.Printf("[Sale] acknowledgment email failed: %v", err)
		}
	}()
}
