package database

import (
    "fmt"
    "franchise-saas-backend/internal/models"
    "log"

    "github.com/spf13/viper"
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
    "gorm.io/gorm/logger"
)

var DB *gorm.DB

func ConnectDB() (*gorm.DB, error) {
    dsn := fmt.Sprintf(
        "host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
        viper.GetString("db_host"),
        viper.GetString("db_user"),
        viper.GetString("db_password"),
        viper.GetString("db_name"),
        viper.GetString("db_port"),
    )

    var err error
    DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
        Logger: logger.Default.LogMode(logger.Info),
    })
    if err != nil {
        return nil, fmt.Errorf("failed to connect database: %w", err)
    }

    log.Println("✅ Database connected")

    // МИГРАЦИЯ
    err = DB.AutoMigrate(
        // Старые модели
        &models.User{},
        &models.Tenant{},
        &models.Salon{},
        &models.Order{},
        &models.Task{},
        &models.Checklist{},
        &models.Plan{},
        &models.Notification{},
        &models.Invoice{},
        
        // === НОВЫЕ МОДЕЛИ (ЭТАП 1: Менеджер Салона) ===
        &models.Lead{},                 // CRM Лиды
        &models.LeadActivity{},         // История контактов
        &models.ChecklistTemplate{},    // Шаблоны чек-листов
        &models.ChecklistTemplateItem{},// Пункты шаблона
        &models.AssignedChecklist{},    // Назначенные чек-листы
        &models.ChecklistResponse{},    // Ответы на чек-листы
    )
    if err != nil {
        return nil, fmt.Errorf("failed to migrate database: %w", err)
    }

    log.Println("✅ Database migrated (new tables added)")
    return DB, nil
}

func GetDB() *gorm.DB {
    return DB
}
