# bochka for Golang

`bochka` is a Golang package designed to streamline your testing environment when working with PostgreSQL. It provides a handy helper that initializes a ready-to-use PostgreSQL instance within Docker. This temporary database instance lasts for the duration of your test run, making it ideal for integration tests or any scenario where a transient database is beneficial.

## Features:
- **Transient PostgreSQL Instances**: Quickly set up a PostgreSQL instance that only lasts for the duration of your tests.
- **Version Control**: Easily specify the desired version of the PostgreSQL image.
- **Seamless Integration with Docker**: No need for complex Docker setups, `bochka` handles it for you.

## Example Usage:
Below is an example showcasing how `bochka` can be used for testing:

```go
func TestDate_WithBD(t *testing.T) {
    helper := bochka.NewPostgreTestHelper(t, bochka.WithTimeout(10*time.Second))
    helper.Run("14.5")
	
    t.Cleanup(func() {
    _, err := helper.Pool.Exec(helper.Context, `DROP TABLE IF EXISTS tmp1`)
        if err != nil {
			t.Error("Test table deletion failed:", err)
        }
        helper.Close()
    })

	_, err := helper.Pool.Exec(ctx, `
CREATE TABLE IF NOT EXISTS tmp1 (
	testDate date
); `)
	if err != nil {
		t.Error("Test table creation failed:", err)
	}

	t.Run("test date 1", func(t *testing.T) {
		inputDate := Now()
		var returnedDate Date
		err = helper.Pool.QueryRow(helper.Context, `
INSERT INTO tmp1(testdate) VALUES($1) RETURNING testdate;
`, inputDate).Scan(&returnedDate)
		if err != nil {
			t.Error("INSERT query failed:", err)
		}
		if returnedDate == inputDate {
			t.Log("success!")
			t.Logf("returned inputDate: '%s'", returnedDate)
			t.Logf("input inputDate: '%s'", inputDate)
		} else {
			t.Errorf("dates are not equal, have '%s', want '%s'", returnedDate, inputDate)
		}
	})
}
```

## Parameters:
- **SetupPostgreTestHelper(t, version string)**:
    - `t`: The test handler.
    - `version`: The version of the PostgreSQL image you would like to use (e.g., "14.5").

## Installation:

To install `bochka`, use the typical `go get`:

```bash
go get -u github.com/kaatinga/bochka
```

## Contributing:
If you would like to contribute to `bochka`, please raise an issue or submit a pull request on our GitHub repository.
