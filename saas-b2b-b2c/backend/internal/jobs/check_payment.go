package jobs

import (
    "context"
    "fmt"
    "log"
    "time"

    "franchise-saas-backend/internal/models"
    "franchise-saas-backend/internal/services"
)

type PaymentJob struct {
    tenantService *services.AdminService
    notifService  *services.NotificationService
}

func NewPaymentJob(ts *services.AdminService, ns *services.NotificationService) *PaymentJob {
    return &PaymentJob{tenantService: ts, notifService: ns}
}

func (j *PaymentJob) Run() {
    log.Println("Running Payment Check Job...")
    ctx := context.Background()

    // Передаем ctx в функцию
    tenants, err := j.tenantService.GetAllTenants(ctx)
    if err != nil {
        log.Printf("Error fetching tenants: %v", err)
        return
    }

    now := time.Now()

    for _, t := range tenants {
        if t.PaidUntil == nil {
            continue
        }

        hoursLeft := t.PaidUntil.Sub(now).Hours()
        daysLeft := int(hoursLeft / 24)

        if daysLeft >= 0 && daysLeft <= 3 {
            msg := fmt.Sprintf("До окончания подписки осталось %d дней. Пожалуйста, продлите оплату.", daysLeft)
            err := j.notifService.CreateNotification(ctx, t.ID, models.NotificationTypePayment, "Окончание подписки", msg)
            if err != nil {
                log.Printf("Error creating notification for tenant %s: %v", t.ID, err)
            }
        }

        if daysLeft < 0 {
            err := j.notifService.CreateNotification(ctx, t.ID, models.NotificationTypePayment, "Подписка просрочена", "Ваша подписка просрочена. Оплатите для возобновления доступа.")
            if err != nil {
                log.Printf("Error creating overdue notification for tenant %s: %v", t.ID, err)
            }
        }
    }
}
