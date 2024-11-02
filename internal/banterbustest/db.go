package banterbustest

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"runtime"

	"github.com/google/uuid"
	"github.com/pressly/goose/v3"

	sqlc "gitlab.com/hmajid2301/banterbus/internal/store/db"

	// used to connect to sqlite
	_ "modernc.org/sqlite"
)

func CreateDB(_ context.Context) (*sql.DB, error) {
	db, err := sql.Open("sqlite", "file::memory:?cache=shared")
	if err != nil {
		return db, err
	}

	err = runDBMigrations(db)
	if err != nil {
		return db, err
	}

	err = FillWithDummyData(db)
	return db, err
}

func runDBMigrations(db *sql.DB) error {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return fmt.Errorf("failed to get current filename")
	}

	dir := path.Join(path.Dir(filename), "..")
	migrationFolder := filepath.Join(dir, "../")

	fs := os.DirFS(migrationFolder)
	goose.SetBaseFS(fs)

	if err := goose.SetDialect("sqlite3"); err != nil {
		return err
	}

	if err := goose.Up(db, "db/migrations"); err != nil {
		return err
	}
	return nil
}

func FillWithDummyData(db *sql.DB) error {
	tx, err := db.Begin()
	ctx := context.Background()
	if err != nil {
		return err
	}

	queries := sqlc.New(db)
	groups := []string{
		"programming_group",
		"cat_group",
		"bike_group",
		"horse_group",
		"colour_group",
		"animal",
	}

	groupNameToID := map[string]map[string]string{}

	for _, group := range groups {
		groupNameToID[group] = map[string]string{}
		questionGroup, err := queries.WithTx(tx).AddQuestionsGroup(ctx, sqlc.AddQuestionsGroupParams{
			ID:        uuid.Must(uuid.NewV7()).String(),
			GroupName: group,
			GroupType: "questions",
		})
		if err != nil {
			return err
		}
		groupNameToID[group]["questions"] = questionGroup.ID

		answerGroup, err := queries.WithTx(tx).AddQuestionsGroup(ctx, sqlc.AddQuestionsGroupParams{
			ID:        uuid.Must(uuid.NewV7()).String(),
			GroupName: group,
			GroupType: "answers",
		})
		if err != nil {
			return err
		}

		groupNameToID[group]["answers"] = answerGroup.ID
	}

	questions := []struct {
		GameName  string
		Round     string
		Enabled   bool
		Question  string
		Language  string
		GroupName string
		GroupType string
	}{
		{"fibbing_it", "likely", false, "to get arrested", "en-GB", "", ""},
		{"fibbing_it", "likely", true, "to eat ice-cream from the tub", "en-GB", "", ""},
		{"fibbing_it", "likely", true, "to fight a police person", "en-GB", "", ""},
		{"fibbing_it", "likely", true, "to fight a horse", "en-GB", "", ""},
		{
			"fibbing_it",
			"free_form",
			true,
			"What do you think about programmers?",
			"en-GB",
			"programming_group",
			"questions",
		},
		{
			"fibbing_it",
			"free_form",
			true,
			"What don't you like about programmers?",
			"en-GB",
			"programming_group",
			"questions",
		},
		{
			"fibbing_it",
			"free_form",
			true,
			"what don't you think about programmers?",
			"en-GB",
			"programming_group",
			"questions",
		},
		{"fibbing_it", "free_form", true, "what dont you think about cats", "en-GB", "cat_group", "questions"},
		{"fibbing_it", "free_form", true, "what don't you like about cats?", "en-GB", "cat_group", "questions"},
		{"fibbing_it", "free_form", false, "what do you like about cats?", "en-GB", "cat_group", "questions"},
		{"fibbing_it", "free_form", true, "what do you think about cats", "en-GB", "cat_group", "questions"},
		{"fibbing_it", "free_form", true, "A funny question?", "en-GB", "bike_group", "questions"},
		{"fibbing_it", "free_form", true, "Favourite bike colour?", "en-GB", "bike_group", "questions"},
		{"fibbing_it", "opinion", true, "lame", "en-GB", "horse_group", "answers"},
		{"fibbing_it", "opinion", true, "tasty", "en-GB", "horse_group", "answers"},
		{"fibbing_it", "opinion", true, "cool", "en-GB", "horse_group", "answers"},
		{"fibbing_it", "opinion", true, "What do you think about camels?", "en-GB", "horse_group", "questions"},
		{"fibbing_it", "opinion", true, "What do you think about horses?", "en-GB", "horse_group", "questions"},
		{"fibbing_it", "opinion", true, "purple", "en-GB", "colour_group", "answers"},
		{"fibbing_it", "opinion", true, "blue", "en-GB", "colour_group", "answers"},
		{"fibbing_it", "opinion", true, "red", "en-GB", "colour_group", "answers"},
		{"fibbing_it", "opinion", true, "What is your favourite colour?", "en-GB", "colour_group", "questions"},
		{"fibbing_it", "opinion", true, "What is your least favourite colour?", "en-GB", "colour_group", "questions"},
		{"fibbing_it", "opinion", true, "Strongly Agree", "en-GB", "animal_group", "answers"},
		{"fibbing_it", "opinion", true, "Agree", "en-GB", "animal_group", "answers"},
		{"fibbing_it", "opinion", true, "Disagree", "en-GB", "animal_group", "answers"},
		{"fibbing_it", "opinion", true, "Are cats cute?", "en-GB", "animal_group", "questions"},
		{"fibbing_it", "opinion", true, "Dogs are cuter than cats?", "en-GB", "animal_group", "questions"},
	}

	for _, q := range questions {
		groupID := groupNameToID[q.GroupName][q.GroupType]

		_, err := queries.WithTx(tx).AddQuestion(ctx, sqlc.AddQuestionParams{
			ID:           uuid.Must(uuid.NewV7()).String(),
			GameName:     q.GameName,
			Round:        q.Round,
			Question:     q.Question,
			LanguageCode: q.Language,
			GroupID:      groupID,
		})
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}
