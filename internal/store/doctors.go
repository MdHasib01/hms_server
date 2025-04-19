package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
)

type Availability struct {
	Day       string `json:"day"`
	StartTime string `json:"startTime"`
	EndTime   string `json:"endTime"`
}
type Doctor struct {
	FirstName     string         `json:"firstname"`
	LastName      string         `json:"lastname"`
	UserName      string         `json:"username"`
	Password      string         `json:"-"`
	Age           string         `json:"age"`
	Gender        string         `json:"gender"`
	Email         string         `json:"email"`
	Mobile        string         `json:"mobile"`
	MaritalStatus string         `json:"marital_status"`
	Designation   string         `json:"designation"`
	Qualification string         `json:"qualification"`
	BloodGroup    string         `json:"bloodGroup"`
	Address       string         `json:"address"`
	Country       string         `json:"country"`
	State         string         `json:"state"`
	City          string         `json:"city"`
	PostalCode    string         `json:"postalCode"`
	Availability  []Availability `json:"availability"`
}

type DoctorStore struct {
	db *sql.DB
}

func (s *DoctorStore) GetByID(ctx context.Context, id int64) (*Doctor, error) {
	query := `
        SELECT 
    d.id,
    d.first_name AS firstname,
    d.last_name AS lastname,
    d.username,
    d.age,
    d.gender,
    d.email,
    d.mobile,
    d.marital_status,
    d.designation,
    d.qualification,
    d.blood_group AS "bloodGroup",
    d.address,
    d.country,
    d.state,
    d.city,
    d.postal_code AS "postalCode",
    COALESCE(
        json_agg(
            json_build_object(
                'day', a.day,
                'startTime', a.start_time,
                'endTime', a.end_time
            )
        ) FILTER (WHERE a.id IS NOT NULL),
        '[]'::json
    ) AS availability
FROM 
    doctors d
LEFT JOIN 
    availability a ON d.id = a.doctor_id
WHERE 
    d.id = $1  -- Replace $1 with the doctor ID you want to retrieve
GROUP BY 
    d.id, d.first_name, d.last_name, d.username, d.age, d.gender, d.email, 
    d.mobile, d.marital_status, d.designation, d.qualification, d.blood_group,
    d.address, d.country, d.state, d.city, d.postal_code;`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	var doctor = &Doctor{}
	var availabilityJSON []byte
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&doctor.FirstName,
		&doctor.LastName,
		&doctor.UserName,
		&doctor.Password,
		&doctor.Age,
		&doctor.Gender,
		&doctor.Email,
		&doctor.Mobile,
		&doctor.MaritalStatus,
		&doctor.Designation,
		&doctor.Qualification,
		&doctor.BloodGroup,
		&doctor.Address,
		&doctor.Country,
		&doctor.State,
		&doctor.City,
		&doctor.PostalCode,
		&availabilityJSON,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}
	if err := json.Unmarshal(availabilityJSON, &doctor.Availability); err != nil {
		return nil, fmt.Errorf("error parsing availability data: %w", err)
	}
	return doctor, nil
}

type DoctorFilter struct {
	Designation   string
	Qualification string
	Country       string
	State         string
	City          string
	Day           string // Availability day
}

// GetDoctorsByFilter returns doctors filtered by specified criteria
func (s *DoctorStore) GetDoctorsByFilter(ctx context.Context, filter DoctorFilter) ([]*Doctor, error) {
	// Start building the query with base selection
	baseQuery := `
        SELECT 
           
            d.first_name AS firstname,
            d.last_name AS lastname,
            d.username,
            d.password,
            d.age,
            d.gender,
            d.email,
            d.mobile,
            d.marital_status,
            d.designation,
            d.qualification,
            d.blood_group AS "bloodGroup",
            d.address,
            d.country,
            d.state,
            d.city,
            d.postal_code AS "postalCode",
            COALESCE(
                json_agg(
                    json_build_object(
                        'day', a.day,
                        'startTime', a.start_time,
                        'endTime', a.end_time
                    )
                ) FILTER (WHERE a.id IS NOT NULL),
                '[]'::json
            ) AS availability
        FROM 
            doctors d
        LEFT JOIN 
            availability a ON d.id = a.doctor_id`

	// Add where clause conditionally
	whereConditions := []string{}
	args := []interface{}{}
	argPos := 1

	if filter.Designation != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("d.designation = $%d", argPos))
		args = append(args, filter.Designation)
		argPos++
	}

	if filter.Qualification != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("d.qualification = $%d", argPos))
		args = append(args, filter.Qualification)
		argPos++
	}

	if filter.Country != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("d.country = $%d", argPos))
		args = append(args, filter.Country)
		argPos++
	}

	if filter.State != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("d.state = $%d", argPos))
		args = append(args, filter.State)
		argPos++
	}

	if filter.City != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("d.city = $%d", argPos))
		args = append(args, filter.City)
		argPos++
	}

	if filter.Day != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("a.day = $%d", argPos))
		args = append(args, filter.Day)
		argPos++
	}

	// Construct the final query
	query := baseQuery
	if len(whereConditions) > 0 {
		query += " WHERE " + whereConditions[0]
		for i := 1; i < len(whereConditions); i++ {
			query += " AND " + whereConditions[i]
		}
	}

	// Add group by and order by
	query += `
        GROUP BY 
            d.id, d.first_name, d.last_name, d.username, d.password, d.age, d.gender, d.email, 
            d.mobile, d.marital_status, d.designation, d.qualification, d.blood_group,
            d.address, d.country, d.state, d.city, d.postal_code
        ORDER BY
            d.id;`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	doctors := []*Doctor{}

	for rows.Next() {
		var doctor Doctor
		var availabilityJSON []byte

		err := rows.Scan(
			&doctor.FirstName,
			&doctor.LastName,
			&doctor.UserName,
			&doctor.Password,
			&doctor.Age,
			&doctor.Gender,
			&doctor.Email,
			&doctor.Mobile,
			&doctor.MaritalStatus,
			&doctor.Designation,
			&doctor.Qualification,
			&doctor.BloodGroup,
			&doctor.Address,
			&doctor.Country,
			&doctor.State,
			&doctor.City,
			&doctor.PostalCode,
			&availabilityJSON,
		)
		if err != nil {
			return nil, err
		}

		// Parse the JSON availability data into the slice
		if err := json.Unmarshal(availabilityJSON, &doctor.Availability); err != nil {
			return nil, fmt.Errorf("error parsing availability data: %w", err)
		}

		doctors = append(doctors, &doctor)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return doctors, nil
}

// GetDoctorsByAvailability finds doctors available on a specific day and time
func (s *DoctorStore) GetDoctorsByAvailability(ctx context.Context, day string, time string) ([]*Doctor, error) {
	query := `
        SELECT 
            d.id,
            d.first_name AS firstname,
            d.last_name AS lastname,
            d.username,
            d.password,
            d.age,
            d.gender,
            d.email,
            d.mobile,
            d.marital_status,
            d.designation,
            d.qualification,
            d.blood_group AS "bloodGroup",
            d.address,
            d.country,
            d.state,
            d.city,
            d.postal_code AS "postalCode",
            COALESCE(
                json_agg(
                    json_build_object(
                        'day', a.day,
                        'startTime', a.start_time,
                        'endTime', a.end_time
                    )
                ) FILTER (WHERE a.id IS NOT NULL),
                '[]'::json
            ) AS availability
        FROM 
            doctors d
        LEFT JOIN 
            availability a ON d.id = a.doctor_id
        WHERE 
            a.day = $1 AND $2::time BETWEEN a.start_time::time AND a.end_time::time
        GROUP BY 
            d.id, d.first_name, d.last_name, d.username, d.password, d.age, d.gender, d.email, 
            d.mobile, d.marital_status, d.designation, d.qualification, d.blood_group,
            d.address, d.country, d.state, d.city, d.postal_code
        ORDER BY
            d.id;`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	rows, err := s.db.QueryContext(ctx, query, day, time)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	doctors := []*Doctor{}

	for rows.Next() {
		var doctor Doctor
		var availabilityJSON []byte

		err := rows.Scan(
			&doctor.FirstName,
			&doctor.LastName,
			&doctor.UserName,
			&doctor.Password,
			&doctor.Age,
			&doctor.Gender,
			&doctor.Email,
			&doctor.Mobile,
			&doctor.MaritalStatus,
			&doctor.Designation,
			&doctor.Qualification,
			&doctor.BloodGroup,
			&doctor.Address,
			&doctor.Country,
			&doctor.State,
			&doctor.City,
			&doctor.PostalCode,
			&availabilityJSON,
		)
		if err != nil {
			return nil, err
		}

		// Parse the JSON availability data into the slice
		if err := json.Unmarshal(availabilityJSON, &doctor.Availability); err != nil {
			return nil, fmt.Errorf("error parsing availability data: %w", err)
		}

		doctors = append(doctors, &doctor)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return doctors, nil
}
