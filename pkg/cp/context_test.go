package cp

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theopenlane/core/pkg/models"
)

// TestWithValue tests the WithValue helper function
func TestWithValue_String(t *testing.T) {
	ctx := context.Background()
	value := "test-organization-id"

	enrichedCtx := WithValue(ctx, value)

	retrieved := GetValue[string](enrichedCtx)
	require.True(t, retrieved.IsPresent())
	assert.Equal(t, value, retrieved.MustGet())
}

func TestWithValue_Integer(t *testing.T) {
	ctx := context.Background()
	value := int64(1024)

	enrichedCtx := WithValue(ctx, value)

	retrieved := GetValue[int64](enrichedCtx)
	require.True(t, retrieved.IsPresent())
	assert.Equal(t, value, retrieved.MustGet())
}

func TestWithValue_OrgModule(t *testing.T) {
	ctx := context.Background()
	module := models.CatalogTrustCenterModule

	enrichedCtx := WithValue(ctx, module)

	retrieved := GetValue[models.OrgModule](enrichedCtx)
	require.True(t, retrieved.IsPresent())
	assert.Equal(t, module, retrieved.MustGet())
}

func TestWithValue_Struct(t *testing.T) {
	type TestConfig struct {
		Timeout int
		Retries int
	}

	ctx := context.Background()
	config := TestConfig{Timeout: 30, Retries: 3}

	enrichedCtx := WithValue(ctx, config)

	retrieved := GetValue[TestConfig](enrichedCtx)
	require.True(t, retrieved.IsPresent())
	assert.Equal(t, config, retrieved.MustGet())
}

func TestWithValue_MultipleTypes(t *testing.T) {
	ctx := context.Background()

	// Add multiple different types
	ctx = WithValue(ctx, "string-value")
	ctx = WithValue(ctx, 42)
	ctx = WithValue(ctx, models.CatalogComplianceModule)

	// Verify all values can be retrieved with correct types
	strVal := GetValue[string](ctx)
	require.True(t, strVal.IsPresent())
	assert.Equal(t, "string-value", strVal.MustGet())

	intVal := GetValue[int](ctx)
	require.True(t, intVal.IsPresent())
	assert.Equal(t, 42, intVal.MustGet())

	moduleVal := GetValue[models.OrgModule](ctx)
	require.True(t, moduleVal.IsPresent())
	assert.Equal(t, models.CatalogComplianceModule, moduleVal.MustGet())
}

func TestGetValue_NotFound(t *testing.T) {
	ctx := context.Background()

	// Try to get a value that doesn't exist
	result := GetValue[string](ctx)
	assert.False(t, result.IsPresent())
}

func TestGetValue_WrongType(t *testing.T) {
	ctx := context.Background()
	ctx = WithValue(ctx, 42) // Store an int

	// Try to get it as a string
	result := GetValue[string](ctx)
	assert.False(t, result.IsPresent())
}

func TestTypedContext_ComplexScenario(t *testing.T) {
	// Test a complex scenario that mimics real-world provider resolution
	ctx := context.Background()

	// Build context like how the objects service does
	orgID := "org-123"
	module := models.CatalogTrustCenterModule
	feature := "evidence"

	// Enrich context step by step
	ctx = WithValue(ctx, orgID)
	ctx = WithValue(ctx, module)
	ctx = WithValue(ctx, feature)

	// Verify all values are retrievable
	retrievedFeature := GetValue[string](ctx)
	require.True(t, retrievedFeature.IsPresent())
	assert.Equal(t, feature, retrievedFeature.MustGet()) // Note: GetValue gets the last string value added

	retrievedModule := GetValue[models.OrgModule](ctx)
	require.True(t, retrievedModule.IsPresent())
	assert.Equal(t, module, retrievedModule.MustGet())

	// Note: GetValue[string] returns the last string value added, which is "evidence"
	// In real usage, different types would be used to avoid conflicts
}

func TestTypedContext_RuleMatching(t *testing.T) {
	// Test scenario that mimics how rules would use the context
	ctx := context.Background()
	ctx = WithValue(ctx, models.CatalogTrustCenterModule)
	ctx = WithValue(ctx, "evidence") // feature

	// Simulate rule condition checking using the cleaner API
	moduleMatches := GetValueEquals(ctx, models.CatalogTrustCenterModule)
	featureMatches := GetValueEquals(ctx, "evidence")

	assert.True(t, moduleMatches)
	assert.True(t, featureMatches)

	// Test negative cases
	wrongModuleMatches := GetValueEquals(ctx, models.CatalogComplianceModule)
	wrongFeatureMatches := GetValueEquals(ctx, "policies")

	assert.False(t, wrongModuleMatches)
	assert.False(t, wrongFeatureMatches)
}

func TestGetValueEquals(t *testing.T) {
	ctx := context.Background()

	// Test with string values
	ctx = WithValue(ctx, "test-value")
	assert.True(t, GetValueEquals(ctx, "test-value"))
	assert.False(t, GetValueEquals(ctx, "other-value"))

	// Test with integer values
	ctx = WithValue(ctx, 42)
	assert.True(t, GetValueEquals(ctx, 42))
	assert.False(t, GetValueEquals(ctx, 99))

	// Test with enum values
	ctx = WithValue(ctx, models.CatalogTrustCenterModule)
	assert.True(t, GetValueEquals(ctx, models.CatalogTrustCenterModule))
	assert.False(t, GetValueEquals(ctx, models.CatalogComplianceModule))

	// Test with empty context (should return false for non-zero values)
	emptyCtx := context.Background()
	assert.False(t, GetValueEquals(emptyCtx, "any-value"))
	assert.False(t, GetValueEquals(emptyCtx, 42))
	assert.False(t, GetValueEquals(emptyCtx, models.CatalogTrustCenterModule))

	// Test with zero values
	assert.True(t, GetValueEquals(emptyCtx, ""))                   // empty string is zero value
	assert.True(t, GetValueEquals(emptyCtx, 0))                    // 0 is zero value
	assert.True(t, GetValueEquals(emptyCtx, models.OrgModule(""))) // empty enum is zero value
}

func TestWithHintMultipleValues(t *testing.T) {
	ctx := context.Background()

	keyOne := NewHintKey[string]("example.one")
	keyTwo := NewHintKey[string]("example.two")

	ctx = WithHint(ctx, keyOne, "first")
	ctx = WithHint(ctx, keyTwo, "second")

	one := GetHint(ctx, keyOne)
	require.True(t, one.IsPresent())
	assert.Equal(t, "first", one.MustGet())

	two := GetHint(ctx, keyTwo)
	require.True(t, two.IsPresent())
	assert.Equal(t, "second", two.MustGet())
}

func TestHintSetApply(t *testing.T) {
	ctx := context.Background()
	stringKey := NewHintKey[string]("string")
	intKey := NewHintKey[int]("int")

	hintSet := NewHintSet()
	AddHint(hintSet, stringKey, "value")
	AddHint(hintSet, intKey, 42)

	ctx = hintSet.Apply(ctx)

	stringHint := GetHint(ctx, stringKey)
	require.True(t, stringHint.IsPresent())
	assert.Equal(t, "value", stringHint.MustGet())

	intHint := GetHint(ctx, intKey)
	require.True(t, intHint.IsPresent())
	assert.Equal(t, 42, intHint.MustGet())
}
