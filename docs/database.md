# Database Query Usage

# Query Execution

Use db.Exec to execute queries exepecting no return

```go
func CreateFacility(db *sql.DB, facility *models.FacilityRequest) error {
	_, err := db.Exec(queries.CreateFacility, &facility.Name, &facility.Location, &facility.Description,
		&facility.DonatedBy, &facility.ImageUrl)

	return err
}
```

Use db.QueryRow to query for a single row

```go
func GetSinglefacility(db *sql.DB, id int) *models.FacilityResponse {
	var facility models.FacilityResponse
	db.QueryRow(queries.GetSingleFacility, id).Scan(&facility.Id, &facility.Name, &facility.Location,
		&facility.Description, &facility.DonatedBy, &facility.ImageUrl, &facility.CreatedAt)

	return &facility
}
```

Use db.Query to query for multiple rows

```go
func GetFacilitiesForUser(db *sql.DB, userId int) (*[]models.UserFacilityResponse, error) {
	var facilities []models.UserFacilityResponse
	rows, err := db.Query(queries.GetAllFacilitesForUser, userId)

	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var facility models.UserFacilityResponse
		err := rows.Scan(&facility.Id, &facility.Name, &facility.Location, &facility.ImageUrl, &facility.CreatedAt)

		if err != nil {
			return nil, err
		}

		facilities = append(facilities, facility)
	}

	return &facilities, nil
}
```

# Pagination

- Theres a helper function for pagination in helpers/helpers.go called GetDataHandler
