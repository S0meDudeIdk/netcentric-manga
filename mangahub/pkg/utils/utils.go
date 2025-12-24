package utils

import (
	"fmt"
	"strings"
)

// String Validation Functions

// IsMangaDexUUID checks if a string is a valid MangaDex UUID format
func IsMangaDexUUID(s string) bool {
	return len(s) == 36 && strings.Count(s, "-") == 4
}

// ValidateStatus checks if a manga status string is valid
func ValidateStatus(status string) error {
	if status == "" {
		return nil // Empty status is allowed
	}

	validStatuses := []string{"ongoing", "completed", "hiatus", "dropped", "cancelled"}
	for _, validStatus := range validStatuses {
		if status == validStatus {
			return nil
		}
	}
	return fmt.Errorf("invalid status: %s (must be one of: %s)", status, strings.Join(validStatuses, ", "))
}

// ValidateURL checks if a string is a valid HTTP/HTTPS URL
func ValidateURL(url string) error {
	if url == "" {
		return nil // Empty URL is allowed
	}
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return fmt.Errorf("URL must be a valid HTTP/HTTPS URL")
	}
	return nil
}

// String Manipulation Functions

// SanitizeString removes leading and trailing whitespace
func SanitizeString(s string) string {
	return strings.TrimSpace(s)
}

// CleanMangaTitle removes common suffixes from manga titles for better searching
func CleanMangaTitle(title string) string {
	cleanTitle := strings.TrimSpace(title)
	cleanTitle = strings.TrimSuffix(cleanTitle, " (TV)")
	cleanTitle = strings.TrimSuffix(cleanTitle, " (Dub)")
	cleanTitle = strings.TrimSuffix(cleanTitle, " (Sub)")
	return cleanTitle
}

// ContainsIgnoreCase checks if a string contains a substring (case-insensitive)
func ContainsIgnoreCase(str, substr string) bool {
	return strings.Contains(strings.ToLower(str), strings.ToLower(substr))
}

// EqualsIgnoreCase checks if two strings are equal (case-insensitive)
func EqualsIgnoreCase(str1, str2 string) bool {
	return strings.EqualFold(str1, str2)
}

// URL Building Functions

// BuildMangaDexPageURL constructs the full URL for a MangaDex chapter page
func BuildMangaDexPageURL(baseURL, hash, filename string, dataSaver bool) string {
	quality := "data"
	if dataSaver {
		quality = "data-saver"
	}
	return fmt.Sprintf("%s/%s/%s/%s", baseURL, quality, hash, filename)
}

// ExtractIDFromPrefix extracts an ID by removing a prefix (e.g., "mangadex-123" -> "123")
func ExtractIDFromPrefix(fullID, prefix string) string {
	return strings.TrimPrefix(fullID, prefix)
}

// Validation Functions for Manga Data

// ValidateMangaID checks if a manga ID is valid (no spaces or special characters)
func ValidateMangaID(id string) error {
	if id == "" {
		return fmt.Errorf("manga ID is required")
	}
	if strings.Contains(id, " ") {
		return fmt.Errorf("manga ID cannot contain spaces")
	}
	return nil
}

// ValidateStringLength checks if a string's length is within the specified range
func ValidateStringLength(value, fieldName string, maxLength int) error {
	if len(value) > maxLength {
		return fmt.Errorf("%s too long (max %d characters)", fieldName, maxLength)
	}
	return nil
}

// ValidateNonNegative checks if an integer is non-negative
func ValidateNonNegative(value int, fieldName string) error {
	if value < 0 {
		return fmt.Errorf("%s cannot be negative", fieldName)
	}
	return nil
}

// ValidateNonEmpty checks if a string is not empty
func ValidateNonEmpty(value, fieldName string) error {
	if value == "" {
		return fmt.Errorf("%s is required", fieldName)
	}
	return nil
}

// ValidateSliceNotEmpty checks if a slice is not empty
func ValidateSliceNotEmpty(slice []string, fieldName string) error {
	if len(slice) == 0 {
		return fmt.Errorf("at least one %s is required", fieldName)
	}
	return nil
}

// Format Functions

// FormatChapterNumber formats a chapter number string consistently
func FormatChapterNumber(chapterNum string) string {
	return strings.TrimSpace(chapterNum)
}

// DefaultString returns the default value if the string is empty
func DefaultString(value, defaultValue string) string {
	if value == "" {
		return defaultValue
	}
	return value
}

// DefaultInt returns the default value if the integer is zero or negative
func DefaultInt(value, defaultValue int) int {
	if value <= 0 {
		return defaultValue
	}
	return value
}

// Slice Helper Functions

// StringInSlice checks if a string exists in a slice of strings
func StringInSlice(str string, slice []string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}

// RemoveDuplicateStrings removes duplicate strings from a slice
func RemoveDuplicateStrings(slice []string) []string {
	seen := make(map[string]bool)
	result := []string{}
	for _, str := range slice {
		if !seen[str] {
			seen[str] = true
			result = append(result, str)
		}
	}
	return result
}

// FilterEmptyStrings removes empty strings from a slice
func FilterEmptyStrings(slice []string) []string {
	result := []string{}
	for _, str := range slice {
		if str != "" {
			result = append(result, str)
		}
	}
	return result
}
