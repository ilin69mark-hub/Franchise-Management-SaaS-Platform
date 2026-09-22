-- =====================================================
-- 013_add_missing_fk_indexes.up.sql
-- FK-индексы для таблиц без покрытия — seq scan при
-- WHERE salon_id/tenant_id/dealer_id/status в KPI-сервисе.
-- Идемпотентно: CREATE INDEX IF NOT EXISTS.
-- Покрывает P2-1 аудита; 011 уже закрыл products/lost_sales/promotions.
-- =====================================================

-- leads: SQL 005 имел, но GORM — нет; страхуем
CREATE INDEX IF NOT EXISTS idx_leads_salon ON leads(salon_id);
CREATE INDEX IF NOT EXISTS idx_leads_manager ON leads(manager_id);
CREATE INDEX IF NOT EXISTS idx_leads_status ON leads(status);
CREATE INDEX IF NOT EXISTS idx_leads_created_at ON leads(created_at);

-- tasks: ни в 001, ни в GORM не было индекса
CREATE INDEX IF NOT EXISTS idx_tasks_assigned_to ON tasks(assigned_to);
CREATE INDEX IF NOT EXISTS idx_tasks_created_by ON tasks(created_by);
CREATE INDEX IF NOT EXISTS idx_tasks_salon ON tasks(salon_id);
CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status);

-- notifications: создана в 009 без индексов
CREATE INDEX IF NOT EXISTS idx_notifications_tenant ON notifications(tenant_id);
CREATE INDEX IF NOT EXISTS idx_notifications_user ON notifications(user_id);
CREATE INDEX IF NOT EXISTS idx_notifications_is_read ON notifications(is_read);

-- alerts: создана в 009 без индексов
CREATE INDEX IF NOT EXISTS idx_alerts_user ON alerts(user_id);
CREATE INDEX IF NOT EXISTS idx_alerts_tenant ON alerts(tenant_id);
CREATE INDEX IF NOT EXISTS idx_alerts_status ON alerts(status);

-- invoices: GORM-only, без индекса
CREATE INDEX IF NOT EXISTS idx_invoices_tenant ON invoices(tenant_id);
CREATE INDEX IF NOT EXISTS idx_invoices_status ON invoices(status);

-- checklists: GORM-only, без индекса
CREATE INDEX IF NOT EXISTS idx_checklists_tenant ON checklists(tenant_id);
CREATE INDEX IF NOT EXISTS idx_checklists_assigned ON checklists(assigned_to);
CREATE INDEX IF NOT EXISTS idx_checklists_user ON checklists(user_id);

-- dealer_*: 009 создал без индексов
CREATE INDEX IF NOT EXISTS idx_dealer_tasks_dealer ON dealer_tasks(dealer_id);
CREATE INDEX IF NOT EXISTS idx_dealer_tasks_tenant ON dealer_tasks(tenant_id);
CREATE INDEX IF NOT EXISTS idx_dealer_tasks_status ON dealer_tasks(status);
CREATE INDEX IF NOT EXISTS idx_dealer_requests_dealer ON dealer_requests(dealer_id);
CREATE INDEX IF NOT EXISTS idx_dealer_requests_status ON dealer_requests(status);
CREATE INDEX IF NOT EXISTS idx_marketing_budgets_dealer ON marketing_budgets(dealer_id);
CREATE INDEX IF NOT EXISTS idx_marketing_budgets_quarter ON marketing_budgets(quarter);
CREATE INDEX IF NOT EXISTS idx_dealer_expenses_dealer ON dealer_expenses(dealer_id);
CREATE INDEX IF NOT EXISTS idx_dealer_expenses_period ON dealer_expenses(period);

-- orders: дополняем created_by (001 имел только salon/status)
CREATE INDEX IF NOT EXISTS idx_orders_created_by ON orders(created_by);

-- contracts: 006 имел базовые, добавляем lead_id для JOIN
CREATE INDEX IF NOT EXISTS idx_contracts_lead ON contracts(lead_id);

-- checklist_templates/items: для JOIN по tenant/template
CREATE INDEX IF NOT EXISTS idx_checklist_templates_tenant ON checklist_templates(tenant_id);
CREATE INDEX IF NOT EXISTS idx_checklist_template_items_template ON checklist_template_items(template_id);
