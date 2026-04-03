-- =====================================================
-- МИГРАЦИЯ 005: Функционал Менеджера Салона (Основа)
-- =====================================================

-- 1. Таблица Лидов (CRM для менеджера)
-- Позволяет отслеживать воронку продаж: Новый -> Контакт -> Встреча -> Продажа
CREATE TABLE IF NOT EXISTS leads (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    salon_id UUID NOT NULL REFERENCES salons(id),
    manager_id UUID NOT NULL REFERENCES users(id), -- Кто ведет (Менеджер салона)

    -- Данные клиента
    full_name VARCHAR(255) NOT NULL,
    phone VARCHAR(50),
    email VARCHAR(100),
    
    -- Интересы
    interest_product VARCHAR(255), -- Что интересует (Диван, Кресло)
    budget DECIMAL(10, 2),
    
    -- Статус воронки
    status VARCHAR(50) DEFAULT 'new', -- new, contact, meeting, sale, archive
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 2. История взаимодействий (Комменты, звонки, встречи)
CREATE TABLE IF NOT EXISTS lead_activities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    lead_id UUID NOT NULL REFERENCES leads(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id), -- Кто сделал действие
    
    type VARCHAR(50) NOT NULL, -- call, meeting, note, status_change
    description TEXT,
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 3. Шаблоны чек-листов (Создаются Франчайзером/Дилером)
CREATE TABLE IF NOT EXISTS checklist_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID REFERENCES tenants(id), -- NULL если глобальный (от админа)
    created_by UUID NOT NULL REFERENCES users(id),
    
    title VARCHAR(255) NOT NULL,
    description TEXT,
    
    -- Тип назначения: daily, weekly, monthly
    type VARCHAR(50) DEFAULT 'daily',
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 4. Пункты шаблона чек-листа
CREATE TABLE IF NOT EXISTS checklist_template_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    template_id UUID NOT NULL REFERENCES checklist_templates(id) ON DELETE CASCADE,
    
    text VARCHAR(500) NOT NULL, -- Текст задачи
    order_index INT DEFAULT 0, -- Порядок сортировки
    
    -- Тип проверки: checkbox, number, photo
    validation_type VARCHAR(50) DEFAULT 'checkbox',
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 5. Назначение чек-листа на салон/менеджера
CREATE TABLE IF NOT EXISTS assigned_checklists (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    template_id UUID NOT NULL REFERENCES checklist_templates(id),
    salon_id UUID NOT NULL REFERENCES salons(id),
    assigned_to UUID NOT NULL REFERENCES users(id), -- Менеджер
    
    due_date DATE, -- Срок выполнения (если разовое)
    
    status VARCHAR(50) DEFAULT 'active', -- active, completed, overdue
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 6. Ответы по чек-листу (Выполнение пунктов)
CREATE TABLE IF NOT EXISTS checklist_responses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    assigned_checklist_id UUID NOT NULL REFERENCES assigned_checklists(id),
    item_id UUID NOT NULL REFERENCES checklist_template_items(id),
    
    value TEXT, -- Значение (true/false для чекбокса, число, ссылка на фото)
    completed BOOLEAN DEFAULT FALSE,
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- =====================================================
-- ИНДЕКСЫ ДЛЯ СКОРОСТИ
-- =====================================================
CREATE INDEX idx_leads_salon ON leads(salon_id);
CREATE INDEX idx_leads_manager ON leads(manager_id);
CREATE INDEX idx_leads_status ON leads(status);
CREATE INDEX idx_lead_activities_lead ON lead_activities(lead_id);
CREATE INDEX idx_assigned_checklists_user ON assigned_checklists(assigned_to);
