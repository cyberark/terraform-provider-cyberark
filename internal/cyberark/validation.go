package cyberark

import (
    "context"
    "fmt"
    "regexp"

    "github.com/hashicorp/terraform-plugin-framework/types"
)

func ValidateInputField(ctx context.Context, name string, val types.String, minLen, maxLen int, pattern string) error {
    if val.IsNull() || val.IsUnknown() {
        return nil
    }

    str := val.ValueString()
    if len(str) < minLen || len(str) > maxLen {
        return fmt.Errorf("field %q must be between %d and %d characters; got %d", name, minLen, maxLen, len(str))
    }

    matched, err := regexp.MatchString(pattern, str)
    if err != nil {
        return fmt.Errorf("regex pattern error for %q: %s", name, err.Error())
    }

    if !matched {
        return fmt.Errorf("field %q must match the pattern %q; got: %s", name, pattern, str)
    }

    return nil
}