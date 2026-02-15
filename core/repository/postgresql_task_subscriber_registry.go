package repository

import (
	"database/sql"
	"fmt"

	"github.com/keerapon-som/no_noodle_workflow/core/entitites"
	"github.com/keerapon-som/no_noodle_workflow/core/util"
)

func (r *PostgreSQLNoNoodleWorkflow) SaveSubscriber(sessionKey string, healthCheckURL string, task string, processID string, callbackURL string) error {
	// Implement the logic to complete a task in the workflow using the repository
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	exists, err := r.CheckIsProcessTaskCallBackExist(tx, processID, task, callbackURL)
	if err != nil {
		return err
	}

	if exists {
		return fmt.Errorf("subscriber already exists for processID: %s, task: %s, callbackURL: %s", processID, task, callbackURL)
	}

	err = r.addSubscriber(tx, sessionKey, healthCheckURL, task, processID, callbackURL)
	if err != nil {
		return err
	}

	return nil
}

func (r *PostgreSQLNoNoodleWorkflow) CheckIsProcessTaskCallBackExist(tx *sql.Tx, processID string, task string, callbackURL string) (bool, error) {
	query := "SELECT COUNT(*) FROM subscription WHERE process_id = $1 AND task = $2 AND callback_url = $3"
	var count int
	err := tx.QueryRow(query, processID, task, callbackURL).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *PostgreSQLNoNoodleWorkflow) addSubscriber(tx *sql.Tx, sessionKey string, healthCheckURL string, task string, processID string, callbackURL string) error {
	query := "INSERT INTO subscription (session_key, health_check_url, task, process_id, callback_url, create_date) VALUES ($1, $2, $3, $4, $5, $6)"
	_, err := tx.Exec(query, sessionKey, healthCheckURL, task, processID, callbackURL, util.GetCurrentTime())
	return err
}

func (r *PostgreSQLNoNoodleWorkflow) GetSubscriberBySessionKey(sessionKey string) (*entitites.SubscriberRegistry, error) {
	var subscriber entitites.SubscriberRegistry

	query := "SELECT session_key, process_id, task, health_check_url, callback_url, create_date FROM subscription WHERE session_key = $1"
	err := r.db.QueryRow(query, sessionKey).Scan(
		&subscriber.SessionKey,
		&subscriber.ProcessID,
		&subscriber.Task,
		&subscriber.HealthCheckURL,
		&subscriber.CallbackURL,
		&subscriber.CreateDate,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			// not found
			return nil, nil
		}
		// real DB error
		return nil, err
	}

	return &subscriber, nil
}

func (r *PostgreSQLNoNoodleWorkflow) GetAllSubscribers() (*[]entitites.SubscriberRegistry, error) {
	var subscription []entitites.SubscriberRegistry
	query := "SELECT session_key, process_id, task, health_check_url, callback_url, create_date FROM subscription"
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var subscriber entitites.SubscriberRegistry
		err := rows.Scan(&subscriber.SessionKey, &subscriber.ProcessID, &subscriber.Task, &subscriber.HealthCheckURL, &subscriber.CallbackURL, &subscriber.CreateDate)
		if err != nil {
			return nil, err
		}
		subscription = append(subscription, subscriber)
	}

	return &subscription, nil
}

func (r *PostgreSQLNoNoodleWorkflow) RemoveSubscriber(sessionKey string) error {
	query := "DELETE FROM subscription WHERE session_key = $1"
	_, err := r.db.Exec(query, sessionKey)
	return err
}
