package dto_test

import (
	"encoding/json"
	"reflect"
	"sort"
	"strings"
	"testing"

	"milpa/aplication/dto"
)

// supplyRequestDTOKeys is the exact set of keys SupplyRequestDTO must emit.
//
// The list is spelled out instead of derived from the struct on purpose: a test
// that computes the expectation from the same reflection it checks against can
// only ever prove that the code agrees with itself. A rename that changes the
// wire contract has to break this test and force a decision about it.
var supplyRequestDTOKeys = []string{
	"actual_amount",
	"address",
	"amount_unit",
	"buyer_id",
	"created_at",
	"delivery_deadline",
	"description",
	"id",
	"min_amount_per_provider",
	"multiple_providers",
	"product_name",
	"request_deadline",
	"status",
	"total_amount",
	"updated_at",
}

// TestSupplyRequestDTOEmitsTheExactExpectedKeySet marshals the DTO and compares
// the produced keys against the contract above.
//
// Before the rename the struct carried crossed and misspelled keys
// (`amount_measure` for AmountUnit, `numer_units` for NumberOfUnits,
// `Addrres` for Address, `min_amount_provider` for MinAmountPerProvider) and
// `amount_unit` named AmountPerUnit rather than AmountUnit. Nothing caught it:
// Go does not validate json tags and encoding/json neither errors nor warns
// about a key nobody asked for.
func TestSupplyRequestDTOEmitsTheExactExpectedKeySet(t *testing.T) {
	encoded, err := json.Marshal(dto.SupplyRequestDTO{})
	if err != nil {
		t.Fatalf("marshal SupplyRequestDTO: %v", err)
	}

	var decoded map[string]json.RawMessage
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("unmarshal SupplyRequestDTO: %v", err)
	}

	got := make([]string, 0, len(decoded))
	for key := range decoded {
		got = append(got, key)
	}
	sort.Strings(got)

	want := append([]string(nil), supplyRequestDTOKeys...)
	sort.Strings(want)

	if !reflect.DeepEqual(got, want) {
		t.Errorf("SupplyRequestDTO marshalled keys mismatch\n got: %v\nwant: %v", got, want)
	}
}

// TestSupplyRequestDTOStructDeclaresNoDuplicateJSONKey covers the failure mode
// the rename was one keystroke away from: AmountUnit takes json:"amount_unit",
// which is exactly what AmountPerUnit used to claim. Renaming only one of the
// two would have compiled, passed every test, and silently dropped a field
// from every response, because encoding/json keeps the first of two same-named
// fields and discards the rest without complaint.
func TestSupplyRequestDTOStructDeclaresNoDuplicateJSONKey(t *testing.T) {
	assertNoDuplicateJSONKeys(t, reflect.TypeOf(dto.SupplyRequestDTO{}))
}

// TestSupplyUpdateDTOsDeclareNoDuplicateJSONKey extends the same guard to the
// two update DTOs, which repeat most of the same fields and were renamed in the
// same pass.
func TestSupplyUpdateDTOsDeclareNoDuplicateJSONKey(t *testing.T) {
	assertNoDuplicateJSONKeys(t, reflect.TypeOf(dto.SupplyUpdateAmountsDTO{}))
	assertNoDuplicateJSONKeys(t, reflect.TypeOf(dto.SupplyGeneralUpdateDTO{}))
	assertNoDuplicateJSONKeys(t, reflect.TypeOf(dto.SupplyUpdateTimeDTO{}))
}

func assertNoDuplicateJSONKeys(t *testing.T, typ reflect.Type) {
	t.Helper()

	owners := make(map[string][]string, typ.NumField())
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		if !field.IsExported() {
			continue
		}

		key := strings.Split(field.Tag.Get("json"), ",")[0]
		if key == "" || key == "-" {
			continue
		}
		owners[key] = append(owners[key], field.Name)
	}

	for key, fields := range owners {
		if len(fields) > 1 {
			t.Errorf("%s declares json:%q on %d fields (%s); encoding/json would keep only the first and silently drop the rest",
				typ.Name(), key, len(fields), strings.Join(fields, ", "))
		}
	}
}
