package repo

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	agbatch "github.com/aif-go/ag-core/contribute/agbatch"
)

// SqlJobRepository persists job/step executions to a SQL database.
type SqlJobRepository struct {
	db         *sql.DB
	driverName string
	nextJobID  atomic.Int64
	nextStepID atomic.Int64
	initOnce   sync.Once
	initErr    error
}

func NewSqlJobRepository(db *sql.DB, driverName string) *SqlJobRepository {
	return &SqlJobRepository{db: db, driverName: driverName}
}

func (r *SqlJobRepository) NextJobID() int64  { return r.nextJobID.Add(1) }
func (r *SqlJobRepository) NextStepID() int64 { return r.nextStepID.Add(1) }

func (r *SqlJobRepository) init(ctx context.Context) error {
	r.initOnce.Do(func() { r.initErr = r.createTables(ctx) })
	return r.initErr
}

func (r *SqlJobRepository) createTables(ctx context.Context) error {
	jobDDL := `CREATE TABLE IF NOT EXISTS batch_job_execution (
		id BIGINT PRIMARY KEY, job_name VARCHAR(255) NOT NULL, status VARCHAR(32) NOT NULL,
		exit_code VARCHAR(32) DEFAULT '', exit_description TEXT DEFAULT '',
		start_time DATETIME NOT NULL, end_time DATETIME, last_updated DATETIME NOT NULL,
		failure_exceptions TEXT DEFAULT ''
	)`
	stepDDL := `CREATE TABLE IF NOT EXISTS batch_step_execution (
		id BIGINT PRIMARY KEY, job_execution_id BIGINT NOT NULL, step_name VARCHAR(255) NOT NULL,
		status VARCHAR(32) NOT NULL, exit_code VARCHAR(32) DEFAULT '', exit_description TEXT DEFAULT '',
		start_time DATETIME NOT NULL, end_time DATETIME, last_updated DATETIME NOT NULL,
		read_count BIGINT DEFAULT 0, write_count BIGINT DEFAULT 0, skip_count BIGINT DEFAULT 0,
		retry_count BIGINT DEFAULT 0, filter_count BIGINT DEFAULT 0, failure_exceptions TEXT DEFAULT ''
	)`
	if r.driverName == "postgres" {
		jobDDL = `CREATE TABLE IF NOT EXISTS batch_job_execution (
			id BIGINT PRIMARY KEY, job_name VARCHAR(255) NOT NULL, status VARCHAR(32) NOT NULL,
			exit_code VARCHAR(32) DEFAULT '', exit_description TEXT DEFAULT '',
			start_time TIMESTAMP NOT NULL, end_time TIMESTAMP, last_updated TIMESTAMP NOT NULL,
			failure_exceptions TEXT DEFAULT ''
		)`
	}
	if _, err := r.db.ExecContext(ctx, jobDDL); err != nil { return err }
	if _, err := r.db.ExecContext(ctx, stepDDL); err != nil { return err }
	return nil
}

func (r *SqlJobRepository) SaveJobExecution(ctx context.Context, exec *agbatch.JobExecution) error {
	if err := r.init(ctx); err != nil { return err }
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO batch_job_execution (id, job_name, status, exit_code, exit_description, start_time, end_time, last_updated, failure_exceptions) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		exec.ID, exec.JobName, string(exec.Status), exitCode(exec.ExitStatus), exitDesc(exec.ExitStatus),
		exec.StartTime, nullableTime(exec.EndTime), exec.LastUpdated, errsJSON(exec.FailureExcs))
	return err
}

func (r *SqlJobRepository) UpdateJobExecution(ctx context.Context, exec *agbatch.JobExecution) error {
	if err := r.init(ctx); err != nil { return err }
	_, err := r.db.ExecContext(ctx,
		`UPDATE batch_job_execution SET status=?, exit_code=?, exit_description=?, end_time=?, last_updated=?, failure_exceptions=? WHERE id=?`,
		string(exec.Status), exitCode(exec.ExitStatus), exitDesc(exec.ExitStatus),
		nullableTime(exec.EndTime), exec.LastUpdated, errsJSON(exec.FailureExcs), exec.ID)
	return err
}

func (r *SqlJobRepository) GetJobExecution(ctx context.Context, id int64) (*agbatch.JobExecution, error) {
	if err := r.init(ctx); err != nil { return nil, err }
	row := r.db.QueryRowContext(ctx, `SELECT id, job_name, status, exit_code, exit_description, start_time, end_time, last_updated, failure_exceptions FROM batch_job_execution WHERE id=?`, id)
	exec := &agbatch.JobExecution{}
	var exitCode, exitDesc, failJSON sql.NullString; var endTime sql.NullTime
	if err := row.Scan(&exec.ID, &exec.JobName, (*string)(&exec.Status), &exitCode, &exitDesc, &exec.StartTime, &endTime, &exec.LastUpdated, &failJSON); err != nil {
		if err == sql.ErrNoRows { return nil, nil }; return nil, err
	}
	if exitCode.Valid { exec.ExitStatus = &agbatch.ExitStatus{Code: exitCode.String, Description: exitDesc.String} }
	if endTime.Valid { exec.EndTime = endTime.Time }
	exec.Context = agbatch.NewExecutionContext()
	return exec, nil
}

func (r *SqlJobRepository) SaveStepExecution(ctx context.Context, exec *agbatch.StepExecution) error {
	if err := r.init(ctx); err != nil { return err }
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO batch_step_execution (id, job_execution_id, step_name, status, exit_code, exit_description, start_time, end_time, last_updated, read_count, write_count, skip_count, retry_count, filter_count, failure_exceptions) VALUES (?, 0, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		exec.ID, exec.StepName, string(exec.Status), exitCode(exec.ExitStatus), exitDesc(exec.ExitStatus),
		exec.StartTime, nullableTime(exec.EndTime), exec.LastUpdated, exec.ReadCount, exec.WriteCount, exec.SkipCount, exec.RetryCount, exec.FilterCount, errsJSON(exec.FailureExcs))
	return err
}

func (r *SqlJobRepository) UpdateStepExecution(ctx context.Context, exec *agbatch.StepExecution) error {
	if err := r.init(ctx); err != nil { return err }
	_, err := r.db.ExecContext(ctx,
		`UPDATE batch_step_execution SET status=?, exit_code=?, exit_description=?, end_time=?, last_updated=?, read_count=?, write_count=?, skip_count=?, retry_count=?, filter_count=?, failure_exceptions=? WHERE id=?`,
		string(exec.Status), exitCode(exec.ExitStatus), exitDesc(exec.ExitStatus), nullableTime(exec.EndTime), exec.LastUpdated,
		exec.ReadCount, exec.WriteCount, exec.SkipCount, exec.RetryCount, exec.FilterCount, errsJSON(exec.FailureExcs), exec.ID)
	return err
}

func (r *SqlJobRepository) GetStepExecution(ctx context.Context, id int64) (*agbatch.StepExecution, error) {
	if err := r.init(ctx); err != nil { return nil, err }
	row := r.db.QueryRowContext(ctx, `SELECT id, step_name, status, exit_code, exit_description, start_time, end_time, last_updated, read_count, write_count, skip_count, retry_count, filter_count, failure_exceptions FROM batch_step_execution WHERE id=?`, id)
	exec := &agbatch.StepExecution{}
	var exitCode, exitDesc, failJSON sql.NullString; var endTime sql.NullTime
	if err := row.Scan(&exec.ID, &exec.StepName, (*string)(&exec.Status), &exitCode, &exitDesc, &exec.StartTime, &endTime, &exec.LastUpdated, &exec.ReadCount, &exec.WriteCount, &exec.SkipCount, &exec.RetryCount, &exec.FilterCount, &failJSON); err != nil {
		if err == sql.ErrNoRows { return nil, nil }; return nil, err
	}
	if exitCode.Valid { exec.ExitStatus = &agbatch.ExitStatus{Code: exitCode.String, Description: exitDesc.String} }
	if endTime.Valid { exec.EndTime = endTime.Time }
	exec.Context = agbatch.NewExecutionContext()
	return exec, nil
}

func exitCode(es *agbatch.ExitStatus) string { if es == nil { return "" }; return es.Code }
func exitDesc(es *agbatch.ExitStatus) string { if es == nil { return "" }; return es.Description }
func nullableTime(t time.Time) sql.NullTime { if t.IsZero() { return sql.NullTime{} }; return sql.NullTime{Time: t, Valid: true} }
func errsJSON(errs []error) string {
	msgs := make([]string, len(errs)); for i, e := range errs { msgs[i] = e.Error() }; b, _ := json.Marshal(msgs); return string(b)
}
