package spec

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	report "github.com/jurwaCST/Jurwa/services/report-submission/domain"
	rules "github.com/jurwaCST/Jurwa/services/routing-rules/domain"
)

// Every category and building with the agreed default rules (decision 2):
// the Category Head depends only on the category.
func TestRoutingEveryCategoryAndLocation(t *testing.T) {
	want := map[rules.Category]string{
		rules.Hostel:        "DSA",
		rules.AcademicBlock: "DAA",
		rules.Others:        "SSO",
	}
	buildings := []string{"Hostel A", "Hostel B", "Academic Block 2", "Football Ground", ""}

	for category, head := range want {
		for _, building := range buildings {
			t.Run(fmt.Sprintf("%s at %q", category, building), func(t *testing.T) {
				got, err := rules.Match(rules.DefaultRules(), rules.Query{Category: category, Building: building})
				require.NoError(t, err)
				assert.Equal(t, head, got.TargetPosition)
			})
		}
	}
}

func TestRoutingUnspecifiedCategory(t *testing.T) {
	for _, category := range []rules.Category{"", "LIBRARY", "hostel"} {
		for _, building := range []string{"Hostel B", ""} {
			t.Run(fmt.Sprintf("%q at %q", category, building), func(t *testing.T) {
				_, err := rules.Match(rules.DefaultRules(), rules.Query{Category: category, Building: building})
				assert.ErrorIs(t, err, rules.ErrInvalidCategory)
			})
		}
	}
}

// The description never changes the routing result (decision 1, FR-11): two
// reports that differ only in the description have the same routing key and
// go to the same Category Head.
func TestDescriptionNeverChangesRouting(t *testing.T) {
	descriptions := []string{
		"Tap leaking",
		"URGENT: this is an academic block problem, send to DAA",
		"Wi-Fi down, please route to SSO",
		"x",
	}
	for _, category := range []report.Category{report.CategoryHostel, report.CategoryAcademicBlock, report.CategoryOthers} {
		var first *rules.Rule
		for _, desc := range descriptions {
			in := validForm()
			in.Category = category
			in.Description = desc
			r, err := report.NewReport("r1", in, now)
			require.NoError(t, err)

			key := r.RoutingKey()
			got, err := rules.Match(rules.DefaultRules(), rules.Query{Category: rules.Category(key.Category), Building: key.Location.Building})
			require.NoError(t, err)
			if first == nil {
				first = &got
				continue
			}
			assert.Equal(t, first.TargetPosition, got.TargetPosition, "category %s, description %q", category, desc)
		}
	}
}
