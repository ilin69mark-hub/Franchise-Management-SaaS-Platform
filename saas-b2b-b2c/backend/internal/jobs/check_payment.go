package jobs

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/google/uuid"
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

    // Используем отдельный метод для job
    tenants, err := j.tenantService.GetTenantsWithPaidUntil(ctx)
    if err != nil {
        log.Printf("Error fetching tenants: %v", err)
        return
    }

    now := time.Now()

    for _, t := range tenants {
        paidUntil, ok := t["paid_until"].(time.Time)
        if !ok {
            continue
        }

        hoursLeft := paidUntil.Sub(now).Hours()
        daysLeft := int(hoursLeft / 24)
        
        tenantID, ok := t["id"].(string)
        if !ok {
            continue
        }
        
        id, err := uuid.Parse(tenantID)
        if err != nil {
            continue
        }

        if daysLeft >= 0 && daysLeft <= 3 {
            msg := fmt.Sprintf("До окончания подписки осталось %d дней. Пожалуйста, продлите оплату.", daysLeft)
            err := j.notifService.CreateNotification(ctx, id, models.NotificationTypePayment, "Окончание подписки", msg)
            if err != nil {
                log.Printf("Error creating notification for tenant %s: %v", tenantID, err)
            }
        }

        if daysLeft < 0 {
            err := j.notifService.CreateNotification(ctx, id, models.NotificationTypePayment, "Подписка просрочена", "Ваша подписка просрочена. Оплатите для возобновления доступа.")
            if err != nil {
                log.Printf("Error creating overdue notification for tenant %s: %v", id, err)
            }
        }
    }
}
