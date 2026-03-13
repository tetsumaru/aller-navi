package highlightpdf

import (
	"context"
	"fmt"
	"os"

	"cloud.google.com/go/firestore"
)

// GetUserTarget は Firestore の users/{userID} ドキュメントから target フィールドを返します。
func GetUserTarget(ctx context.Context, userID string) (string, error) {
	databaseID := os.Getenv("FIRESTORE_DATABASE_ID")
	if databaseID == "" {
		databaseID = "(default)"
	}
	client, err := firestore.NewClientWithDatabase(ctx, firestore.DetectProjectID, databaseID)
	if err != nil {
		return "", fmt.Errorf("create firestore client: %w", err)
	}
	defer client.Close()

	doc, err := client.Collection("users").Doc(userID).Get(ctx)
	if err != nil {
		return "", fmt.Errorf("get user document: %w", err)
	}

	val, ok := doc.Data()["target"]
	if !ok {
		return "", fmt.Errorf("target field is missing for user %q", userID)
	}
	target, ok := val.(string)
	if !ok {
		return "", fmt.Errorf("target field is not a string for user %q", userID)
	}

	return target, nil
}
