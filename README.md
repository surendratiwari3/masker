# Masker

Masker is a Go module for flexible, runtime-configurable data masking. It supports **struct tag-based masking**, **runtime overrides**, **context-based masking**, nested structs, slices, maps, and custom masking strategies.

---

## Features

- **Tag-based masking**: Apply masking strategies directly via struct tags.
- **Runtime overrides**: Masking rules can be overridden at runtime per field.
- **Context-based masking**: Dynamically enable/disable masking or override rules using `context.Context`.
- **Nested structs and slices**: Supports complex structures, arrays, and maps.
- **Custom strategies**: Users can define their own masking strategies.
- **Masked copy**: Generate a masked copy without modifying the original struct.

---

## Installation

```bash
go get github.com/surendratiwari3/masker


### Usage

#### 1. Basic Masking (Tag-Based)

```go
package main

import (
	"fmt"
	"github.com/surendratiwari3/masker"
)

type Customer struct {
	Name       string `mask:"strategy:PII"`
	Email      string `mask:"strategy:PII"`
	Phone      string `mask:"strategy:phone"`
	CardNumber string `mask:"strategy:PCI"`
	DOB        string `mask:"strategy:PHI"`
	Password   string `mask:"strategy:CREDENTIALS"`
}

func main() {
	c := &Customer{
		Name:       "John Doe",
		Email:      "john.doe@example.com",
		Phone:      "9876543210",
		CardNumber: "4111111111111111",
		DOB:        "1990-12-31",
		Password:   "supersecret",
	}

	masker.Mask(c)
	fmt.Printf("%+v\n", c)
}
#### 2. Context-Based Masking

```go
package main

import (
	"context"
	"fmt"
	"github.com/surendratiwari3/masker"
)

type Customer struct {
	Name       string `mask:"strategy:PII"`
	Email      string `mask:"strategy:PII"`
	Phone      string `mask:"strategy:phone"`
	CardNumber string `mask:"strategy:PCI"`
	DOB        string `mask:"strategy:PHI"`
	Password   string `mask:"strategy:CREDENTIALS"`
}

func main() {
	c := &Customer{
		Name:       "John Doe",
		Email:      "john.doe@example.com",
		Phone:      "9876543210",
		CardNumber: "4111111111111111",
		DOB:        "1990-12-31",
		Password:   "supersecret",
	}

	// ----------------- Context with overrides -----------------
	overrides := masker.MaskOverrides{
		"Email":      "none",           // skip masking Email
		"Phone":      "full",           // fully mask Phone
		"CardNumber": "partial",        // partial mask CardNumber
	}
	ctx := context.Background()
	ctx = context.WithValue(ctx, "maskOverrides", overrides)
	ctx = context.WithValue(ctx, "disableMasking", false)

	masker.MaskWithContext(c, ctx)
	fmt.Printf("%+v\n", c)

	// ----------------- Context with masking disabled -----------------
	c2 := *c
	ctxDisabled := context.Background()
	ctxDisabled = context.WithValue(ctxDisabled, "disableMasking", true)

	masker.MaskWithContext(&c2, ctxDisabled)
	fmt.Printf("%+v\n", c2)
}

### Current Default Masking Strategies

| Strategy      | Description                          |
|---------------|--------------------------------------|
| partial       | Shows first 2 and last 2 characters |
| full          | Fully masked with `*`               |
| email         | Masks local part of email           |
| phone         | Masks all but last 4 digits         |
| creditcard    | Masks all but last 4 digits         |
| dob           | Shows only year/month (****-**-DD) |
| password      | Fully masked                        |
| token         | Fully masked                        |
| PII           | Group: partial                      |
| PHI           | Group: dob                          |
| PCI           | Group: creditcard                   |
| CREDENTIALS   | Group: full                         |
| FINANCIAL     | Group: partial                      |
| GDPR          | Group: full                         |
| none          | Explicit no-mask                    |


