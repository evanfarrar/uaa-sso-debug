package models

import (
	"database/sql"
	"sort"
	"time"
)

type ScheduledRulesInterface interface {
	Create(ScheduledRule) (ScheduledRule, error)
	Find(int) (ScheduledRule, error)
	Update(ScheduledRule) (int64, error)
	Destroy(ScheduledRule) (int, error)
	FindAllNonRecurringByServiceBindingGuidAndExecutesAt(string, time.Time) ([]ScheduledRule, error)
	FindAllByServiceBindingGuid(string) ([]ScheduledRule, error)
	CountByServiceBindingGuid(string) (int, error)
	NextScheduledEventByServiceBindingGuid(string) (ScheduledEvent, error)
	FindAllRecurringByServiceBindingGuid(string) ([]ScheduledRule, error)
	EventsExecutingWithin(time.Time, time.Duration) (SortableScheduledEvents, error)
	DeleteAllNonRecurringBefore(time.Time) (int, error)
	DeleteAllByServiceBindingGUID(string) (int, error)
}

type ScheduledRules struct{}

func NewScheduledRulesRepo() ScheduledRules {
	return ScheduledRules{}
}

func (repo ScheduledRules) Create(sr ScheduledRule) (ScheduledRule, error) {
	var epoch time.Time
	if sr.CreatedAt.Unix() == epoch.Unix() {
		sr.CreatedAt = time.Now().Truncate(1 * time.Second).UTC()
	} else {
		sr.CreatedAt = sr.CreatedAt.Truncate(1 * time.Second).UTC()
	}

	sr.ExecutesAt = sr.ExecutesAt.Truncate(1 * time.Second).UTC()
	err := Database().Connection.Insert(&sr)

	return sr, err
}

func (repo ScheduledRules) Update(sr ScheduledRule) (int64, error) {
	rowCount, err := Database().Connection.Update(&sr)

	return rowCount, err
}

func (repo ScheduledRules) Find(id int) (ScheduledRule, error) {
	sr := ScheduledRule{}
	err := Database().Connection.SelectOne(&sr, "SELECT * FROM `scheduled_rules` WHERE `id` = ?", id)
	if err == sql.ErrNoRows {
		err = ErrRecordNotFound
	}

	return sr, err
}

func (repo ScheduledRules) Destroy(sr ScheduledRule) (int, error) {
	count, err := Database().Connection.Delete(&sr)
	return int(count), err
}

func (repo ScheduledRules) FindAllNonRecurringByServiceBindingGuidAndExecutesAt(guid string, executionTime time.Time) ([]ScheduledRule, error) {
	var rules []ScheduledRule
	results, err := Database().Connection.Select(ScheduledRule{},
		"SELECT * FROM `scheduled_rules` WHERE `recurrence` = 0 AND `service_binding_guid` = ? AND `executes_at` = ?", guid, executionTime)
	if err != nil {
		return rules, err
	}

	for _, result := range results {
		rule := result.(*ScheduledRule)
		rules = append(rules, *rule)
	}

	return rules, nil
}

func (repo ScheduledRules) FindAllRecurringByServiceBindingGuid(guid string) ([]ScheduledRule, error) {
	var rules []ScheduledRule
	results, err := Database().Connection.Select(ScheduledRule{},
		"SELECT * FROM `scheduled_rules` WHERE `recurrence` != 0 AND `service_binding_guid` = ?", guid)
	if err != nil {
		return rules, err
	}

	for _, result := range results {
		rule := result.(*ScheduledRule)
		rules = append(rules, *rule)
	}

	return rules, nil
}

func (repo ScheduledRules) FindAllByServiceBindingGuid(guid string) ([]ScheduledRule, error) {
	var rules []ScheduledRule
	results, err := Database().Connection.Select(ScheduledRule{}, "SELECT * FROM `scheduled_rules` WHERE `service_binding_guid` = ?", guid)
	if err != nil {
		return rules, err
	}

	for _, result := range results {
		rule := result.(*ScheduledRule)
		if rule.IsRecurring() {
			rules = append(rules, *rule)
		} else if rule.InFuture() {
			rules = append(rules, *rule)
		}
	}

	return rules, nil
}

func (repo ScheduledRules) CountByServiceBindingGuid(guid string) (int, error) {
	query := "SELECT COUNT(id) FROM `scheduled_rules` WHERE `service_binding_guid` = ? AND `enabled` = true AND (`recurrence` > 0 OR `executes_at` > ?)"
	count, err := Database().Connection.SelectInt(query, guid, time.Now())

	return int(count), err
}

func (repo ScheduledRules) NextScheduledEventByServiceBindingGuid(guid string) (ScheduledEvent, error) {
	rules, err := repo.FindAllByServiceBindingGuid(guid)
	events := make(SortableScheduledEvents, 0)
	for _, rule := range rules {
		if rule.IsDisabled() {
			continue
		}

		event, err := rule.NextEvent()
		if err == nil {
			events = append(events, event)
		}
	}
	sort.Sort(events)

	if len(events) == 0 {
		return ScheduledEvent{}, ErrNoScheduledEventsFound
	}

	return events[0], err
}

func (repo ScheduledRules) EventsExecutingWithin(startTime time.Time, duration time.Duration) (SortableScheduledEvents, error) {
	var events SortableScheduledEvents
	endTime := startTime.Add(duration)

	query := "SELECT * FROM `scheduled_rules` WHERE (`executes_at` >= ? AND `executes_at` <= ? AND `recurrence` = 0) OR (`executes_at` < ? AND `recurrence` != 0 AND `enabled` = true)"
	applicableRules, err := Database().Connection.Select(ScheduledRule{}, query, startTime, endTime, endTime)
	if err != nil {
		return events, err
	}

	for _, rule := range applicableRules {
		for _, event := range rule.(*ScheduledRule).FutureEvents(startTime, duration) {
			events = append(events, event)
		}
	}

	return events, nil
}

func (repo ScheduledRules) DeleteAllNonRecurringBefore(lookupTime time.Time) (int, error) {
	result, err := Database().Connection.Exec("DELETE FROM `scheduled_rules` WHERE `executes_at` < ? AND `recurrence` = 0", lookupTime)
	if err != nil {
		return 0, err
	}
	count, _ := result.RowsAffected()
	return int(count), err
}

func (repo ScheduledRules) DeleteAllByServiceBindingGUID(bindingGUID string) (int, error) {
	result, err := Database().Connection.Exec("DELETE FROM `scheduled_rules` WHERE `service_binding_guid` = ?", bindingGUID)
	if err != nil {
		return 0, err
	}
	count, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}
	return int(count), err
}
