// Package store is the data-access layer: it reads and writes customers and
// repairs using a pgx connection pool.
package store

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"CellphoneRepairBackend/internal/models"
)

// ErrNotFound is returned when a requested row does not exist.
var ErrNotFound = errors.New("not found")

// Store provides database operations backed by a connection pool.
type Store struct {
	pool *pgxpool.Pool
}

// New returns a Store using the given pool.
func New(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

// UpsertCustomer inserts a customer, or updates the first name/email if the phone
// number already exists. Returns the stored row.
func (s *Store) UpsertCustomer(ctx context.Context, c models.Customer) (models.Customer, error) {
	row := s.pool.QueryRow(ctx, `
		insert into customers (phone_number, name, first_name, email)
		values ($1, $2, $3, $4)
		on conflict (phone_number) do update
			set first_name = excluded.first_name,
			    email = excluded.email,
			    updated_at = now()
		returning phone_number, name, first_name, email, created_at, updated_at`,
		c.PhoneNumber, c.Name, c.FirstName, c.Email)

	var out models.Customer
	err := row.Scan(&out.PhoneNumber, &out.Name, &out.FirstName, &out.Email, &out.CreatedAt, &out.UpdatedAt)
	return out, err
}

// GetCustomer looks up a customer by phone number.
func (s *Store) GetCustomer(ctx context.Context, phone string) (models.Customer, error) {
	row := s.pool.QueryRow(ctx, `
		select phone_number, name, first_name, email, created_at, updated_at
		from customers where phone_number = $1`, phone)

	var out models.Customer
	err := row.Scan(&out.PhoneNumber, &out.Name, &out.FirstName, &out.Email, &out.CreatedAt, &out.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.Customer{}, ErrNotFound
	}
	return out, err
}

// CustomerDevices returns a customer's device models with their repair history.
func (s *Store) CustomerDevices(ctx context.Context, phone string) ([]models.CustomerDevice, error) {
	rows, err := s.pool.Query(ctx, `
		select d.id, d.model_name, r.id, r.issue_description, r.status, r.created_at
		from customer_devices d
		left join repairs r on r.customer_device_id = d.id
		where d.customer_phone = $1
		order by d.model_name, r.created_at desc`, phone)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	devices := make([]models.CustomerDevice, 0)
	byID := make(map[int64]int)
	for rows.Next() {
		var device models.CustomerDevice
		var repairID *int64
		var issue, status *string
		var createdAt *time.Time
		if err := rows.Scan(&device.ID, &device.ModelName, &repairID, &issue, &status, &createdAt); err != nil {
			return nil, err
		}
		index, exists := byID[device.ID]
		if !exists {
			device.Repairs = make([]models.DeviceRepair, 0)
			devices = append(devices, device)
			index = len(devices) - 1
			byID[device.ID] = index
		}
		if repairID != nil {
			devices[index].Repairs = append(devices[index].Repairs, models.DeviceRepair{
				ID: *repairID, IssueDescription: *issue, Status: *status, CreatedAt: *createdAt,
			})
		}
	}
	return devices, rows.Err()
}

// GetOrCreateCustomerDevice matches models case-insensitively for one customer.
func (s *Store) GetOrCreateCustomerDevice(ctx context.Context, customerPhone, modelName string) (models.CustomerDevice, error) {
	modelName = strings.Join(strings.Fields(strings.TrimSpace(modelName)), " ")
	modelKey := strings.ToLower(modelName)
	var device models.CustomerDevice
	err := s.pool.QueryRow(ctx, `
		insert into customer_devices (customer_phone, model_name, model_key)
		values ($1, $2, $3)
		on conflict (customer_phone, model_key) do update set model_key = excluded.model_key
		returning id, model_name`, customerPhone, modelName, modelKey).Scan(&device.ID, &device.ModelName)
	return device, err
}

const employeeColumns = `id, email, password_hash, role, active, created_at`

func scanEmployee(row pgx.Row) (models.Employee, error) {
	var e models.Employee
	err := row.Scan(&e.ID, &e.Email, &e.PasswordHash, &e.Role, &e.Active, &e.CreatedAt)
	return e, err
}

// CreateEmployee inserts a new staff account and returns it.
func (s *Store) CreateEmployee(ctx context.Context, email, passwordHash, role string) (models.Employee, error) {
	row := s.pool.QueryRow(ctx, `
		insert into employees (email, password_hash, role)
		values ($1, $2, $3)
		returning `+employeeColumns, email, passwordHash, role)
	return scanEmployee(row)
}

// GetEmployeeByEmail looks up an employee by email (for login).
func (s *Store) GetEmployeeByEmail(ctx context.Context, email string) (models.Employee, error) {
	row := s.pool.QueryRow(ctx, `select `+employeeColumns+` from employees where email = $1`, email)
	e, err := scanEmployee(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.Employee{}, ErrNotFound
	}
	return e, err
}

// GetEmployeeByID looks up an employee by id.
func (s *Store) GetEmployeeByID(ctx context.Context, id int64) (models.Employee, error) {
	row := s.pool.QueryRow(ctx, `select `+employeeColumns+` from employees where id = $1`, id)
	e, err := scanEmployee(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.Employee{}, ErrNotFound
	}
	return e, err
}

// ListEmployees returns all staff accounts, newest first.
func (s *Store) ListEmployees(ctx context.Context) ([]models.Employee, error) {
	rows, err := s.pool.Query(ctx, `select `+employeeColumns+` from employees order by created_at desc`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	employees := make([]models.Employee, 0)
	for rows.Next() {
		e, err := scanEmployee(rows)
		if err != nil {
			return nil, err
		}
		employees = append(employees, e)
	}
	return employees, rows.Err()
}

// UpdateEmployee updates role, active, and/or password_hash. Nil fields are left unchanged.
func (s *Store) UpdateEmployee(ctx context.Context, id int64, role *string, active *bool, passwordHash *string) (models.Employee, error) {
	row := s.pool.QueryRow(ctx, `
		update employees set
			role = coalesce($2, role),
			active = coalesce($3, active),
			password_hash = coalesce($4, password_hash)
		where id = $1
		returning `+employeeColumns, id, role, active, passwordHash)
	e, err := scanEmployee(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.Employee{}, ErrNotFound
	}
	return e, err
}

const repairColumns = `id, customer_phone, customer_device_id, device, issue_description, status,
	price, notes, parts_used, warranty, estimated_completion,
	intake_employee_id, intake_checklist, front_photo_path, back_photo_path, created_at, updated_at`

func scanRepair(row pgx.Row) (models.Repair, error) {
	var r models.Repair
	err := row.Scan(
		&r.ID, &r.CustomerPhone, &r.CustomerDeviceID, &r.Device, &r.IssueDescription, &r.Status,
		&r.Price, &r.Notes, &r.PartsUsed, &r.Warranty, &r.EstimatedCompletion,
		&r.IntakeEmployeeID, &r.IntakeChecklist, &r.FrontPhotoPath, &r.BackPhotoPath, &r.CreatedAt, &r.UpdatedAt,
	)
	return r, err
}

// CreateRepair inserts a new repair and returns the stored row.
func (s *Store) CreateRepair(ctx context.Context, r models.Repair) (models.Repair, error) {
	var checklist any
	if len(r.IntakeChecklist) > 0 {
		encoded, err := json.Marshal(r.IntakeChecklist)
		if err != nil {
			return models.Repair{}, err
		}
		checklist = string(encoded)
	}
	row := s.pool.QueryRow(ctx, `
		insert into repairs (
			customer_phone, customer_device_id, device, issue_description, intake_employee_id, intake_checklist, front_photo_path, back_photo_path
		)
		values ($1, $2, $3, $4, $5, $6::jsonb, $7, $8)
		returning `+repairColumns,
		r.CustomerPhone, r.CustomerDeviceID, r.Device, r.IssueDescription, r.IntakeEmployeeID,
		checklist, r.FrontPhotoPath, r.BackPhotoPath)
	return scanRepair(row)
}

// GetRepair fetches a single repair by id.
func (s *Store) GetRepair(ctx context.Context, id int64) (models.Repair, error) {
	row := s.pool.QueryRow(ctx, `select `+repairColumns+` from repairs where id = $1`, id)
	r, err := scanRepair(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.Repair{}, ErrNotFound
	}
	return r, err
}

// UpdateRepairStatus sets a repair's status and returns the updated row.
func (s *Store) UpdateRepairStatus(ctx context.Context, id int64, status string) (models.Repair, error) {
	row := s.pool.QueryRow(ctx, `
		update repairs set status = $1, updated_at = now()
		where id = $2
		returning `+repairColumns, status, id)
	r, err := scanRepair(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.Repair{}, ErrNotFound
	}
	return r, err
}

// ListRepairs returns repairs, most recent first. When openOnly is true, only
// repairs that are not COMPLETED or CANCELLED are returned (the work queue).
func (s *Store) ListRepairs(ctx context.Context, openOnly bool) ([]models.Repair, error) {
	query := `select ` + repairColumns + ` from repairs`
	if openOnly {
		query += ` where status not in ('COMPLETED', 'CANCELLED')`
	}
	query += ` order by created_at desc`
	return s.queryRepairs(ctx, query)
}

// ListRepairsByPhone returns all repairs for a given customer, newest first.
func (s *Store) ListRepairsByPhone(ctx context.Context, phone string) ([]models.Repair, error) {
	return s.queryRepairs(ctx,
		`select `+repairColumns+` from repairs where customer_phone = $1 order by created_at desc`,
		phone)
}

func (s *Store) queryRepairs(ctx context.Context, query string, args ...any) ([]models.Repair, error) {
	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	repairs := make([]models.Repair, 0)
	for rows.Next() {
		r, err := scanRepair(rows)
		if err != nil {
			return nil, err
		}
		repairs = append(repairs, r)
	}
	return repairs, rows.Err()
}

// CreateAuthorization stores the signed repair authorization captured at intake.
func (s *Store) CreateAuthorization(ctx context.Context, a models.RepairAuthorization) (models.RepairAuthorization, error) {
	row := s.pool.QueryRow(ctx, `
		insert into repair_authorizations (repair_id, terms_text, signature_url, signed_by_name, employee_id)
		values ($1, $2, $3, $4, $5)
		returning id, repair_id, terms_text, signature_url, signed_by_name, employee_id, signed_at`,
		a.RepairID, a.TermsText, a.SignatureURL, a.SignedByName, a.EmployeeID)

	var out models.RepairAuthorization
	err := row.Scan(&out.ID, &out.RepairID, &out.TermsText, &out.SignatureURL,
		&out.SignedByName, &out.EmployeeID, &out.SignedAt)
	return out, err
}

// GetAuthorizationByRepair returns the signed authorization for a repair, if any.
func (s *Store) GetAuthorizationByRepair(ctx context.Context, repairID int64) (models.RepairAuthorization, error) {
	row := s.pool.QueryRow(ctx, `
		select id, repair_id, terms_text, signature_url, signed_by_name, employee_id, signed_at
		from repair_authorizations where repair_id = $1 order by signed_at desc limit 1`, repairID)

	var out models.RepairAuthorization
	err := row.Scan(&out.ID, &out.RepairID, &out.TermsText, &out.SignatureURL,
		&out.SignedByName, &out.EmployeeID, &out.SignedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.RepairAuthorization{}, ErrNotFound
	}
	return out, err
}

// CreateNotification records a message we sent (or tried to send) to a customer.
func (s *Store) CreateNotification(ctx context.Context, customerPhone string, repairID int64, message, status string) error {
	_, err := s.pool.Exec(ctx, `
		insert into notifications (customer_phone, repair_id, type, message, status)
		values ($1, $2, 'STATUS_UPDATE', $3, $4)`,
		customerPhone, repairID, message, status)
	return err
}
