package banterbustest

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand"
	"os"
	"path"
	"runtime"

	// INFO: Driver to connect to postgres to run DB migrations
	_ "github.com/jackc/pgx/v5/stdlib"
	pgxUUID "github.com/vgarvardt/pgx-google-uuid/v5"

	"gitlab.com/hmajid2301/banterbus/internal/store/db"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pressly/goose/v3"
)

func CreateDB(ctx context.Context) (*pgxpool.Pool, error) {
	uri := getURI()
	pool, err := pgxpool.New(ctx, uri)
	if err != nil {
		return pool, fmt.Errorf("failed to get database: %w", err)
	}

	randomNumLimit := 1000000
	dbName := fmt.Sprintf("banterbus_test_%d", rand.Intn(randomNumLimit))
	_, err = pool.Exec(ctx, fmt.Sprintf("CREATE DATABASE %s", dbName))
	if err != nil {
		return pool, err
	}

	fmt.Println("test database name: ", dbName)

	pgxConfig, err := pgxpool.ParseConfig(fmt.Sprintf("%s/%s", uri, dbName))
	if err != nil {
		return pool, fmt.Errorf("failed to parse db uri: %w", err)
	}

	pgxConfig.AfterConnect = func(_ context.Context, conn *pgx.Conn) error {
		pgxUUID.Register(conn.TypeMap())
		return nil
	}
	pool, err = pgxpool.NewWithConfig(ctx, pgxConfig)
	if err != nil {
		return pool, fmt.Errorf("failed to setup database: %w", err)
	}

	err = runDBMigrations(pool)
	if err != nil {
		return pool, err
	}

	err = FillWithDummyData(ctx, pool)
	return pool, err
}

func getURI() string {
	uri := os.Getenv("BANTERBUS_DB_URI")
	if uri == "" {
		uri = "postgresql://postgres:postgres@localhost:5432"
	}
	return uri
}

func RemoveDB(ctx context.Context, pool *pgxpool.Pool) error {
	dbName, err := getDatabaseName(pool)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer pool.Close()

	_, err = pool.Exec(ctx, fmt.Sprintf("DROP DATABASE IF EXISTS %s", dbName))
	if err != nil {
		return fmt.Errorf("failed to drop database: %w", err)
	}

	return nil
}

func getDatabaseName(pool *pgxpool.Pool) (string, error) {
	connConfig := pool.Config().ConnConfig
	connString := connConfig.ConnString()

	config, err := pgx.ParseConfig(connString)
	if err != nil {
		return "", fmt.Errorf("failed to parse connection string: %w", err)
	}

	return config.Database, nil
}

func runDBMigrations(pool *pgxpool.Pool) error {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return fmt.Errorf("failed to get current filename")
	}

	dir := path.Join(path.Dir(filename), "..", "store", "db", "sqlc")

	fs := os.DirFS(dir)
	goose.SetBaseFS(fs)

	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}

	cp := pool.Config().ConnConfig.ConnString()
	sqlDB, err := sql.Open("pgx/v5", cp)
	if err != nil {
		return err
	}

	err = goose.Up(sqlDB, "migrations")
	return err
}

func FillWithDummyData(ctx context.Context, pool *pgxpool.Pool) error {
	tx, err := pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}

	queries := db.New(pool)
	groups := []string{
		"programming_group",
		"cat_group",
		"bike_group",
		"horse_group",
		"colour_group",
		"animal_group",
		"all",
	}

	groupNameToID := map[string]map[string]uuid.UUID{}

	for _, group := range groups {
		groupNameToID[group] = map[string]uuid.UUID{}
		questionGroup, err := queries.WithTx(tx).AddQuestionsGroup(ctx, db.AddQuestionsGroupParams{
			ID:        uuid.Must(uuid.NewV7()),
			GroupName: group,
			GroupType: "questions",
		})
		if err != nil {
			return err
		}
		groupNameToID[group]["questions"] = questionGroup.ID

		answerGroup, err := queries.WithTx(tx).AddQuestionsGroup(ctx, db.AddQuestionsGroupParams{
			ID:        uuid.Must(uuid.NewV7()),
			GroupName: group,
			GroupType: "answers",
		})
		if err != nil {
			return err
		}

		groupNameToID[group]["answers"] = answerGroup.ID
	}

	questions := []struct {
		GameName   string
		QuestionID uuid.UUID
		Round      string
		Enabled    bool
		Question   string
		Locale     string
		GroupName  string
		GroupType  string
	}{
		{"fibbing_it", uuid.New(), "most_likely", false, "to get arrested", "en-GB", "all", "questions"},
		{"fibbing_it", uuid.New(), "most_likely", true, "to eat ice-cream from the tub", "en-GB", "all", "questions"},
		{"fibbing_it", uuid.New(), "most_likely", true, "to fight a police person", "en-GB", "all", "questions"},
		{"fibbing_it", uuid.New(), "most_likely", true, "to fight a horse", "en-GB", "all", "questions"},
		{
			"fibbing_it",
			uuid.New(),
			"free_form",
			true,
			"What do you think about programmers?",
			"en-GB",
			"programming_group",
			"questions",
		},
		{
			"fibbing_it",
			uuid.New(),
			"free_form",
			true,
			"What don't you like about programmers?",
			"en-GB",
			"programming_group",
			"questions",
		},
		{
			"fibbing_it",
			uuid.New(),
			"free_form",
			true,
			"what don't you think about programmers?",
			"en-GB",
			"programming_group",
			"questions",
		},
		{
			"fibbing_it",
			uuid.New(),
			"free_form",
			true,
			"what dont you think about cats",
			"en-GB",
			"cat_group",
			"questions",
		},
		{
			"fibbing_it",
			uuid.New(),
			"free_form",
			true,
			"what don't you like about cats?",
			"en-GB",
			"cat_group",
			"questions",
		},
		{
			"fibbing_it",
			uuid.New(),
			"free_form",
			false,
			"what do you like about cats?",
			"en-GB",
			"cat_group",
			"questions",
		},
		{
			"fibbing_it",
			uuid.New(),
			"free_form",
			true,
			"what do you think about cats",
			"en-GB",
			"cat_group",
			"questions",
		},
		{"fibbing_it", uuid.New(), "free_form", true, "A funny question?", "en-GB", "bike_group", "questions"},
		{"fibbing_it", uuid.New(), "free_form", true, "Favourite bike colour?", "en-GB", "bike_group", "questions"},
		{"fibbing_it", uuid.New(), "multiple_choice", true, "lame", "en-GB", "horse_group", "answers"},
		{"fibbing_it", uuid.New(), "multiple_choice", true, "tasty", "en-GB", "horse_group", "answers"},
		{"fibbing_it", uuid.New(), "multiple_choice", true, "cool", "en-GB", "horse_group", "answers"},
		{
			"fibbing_it",
			uuid.New(),
			"multiple_choice",
			true,
			"What do you think about camels?",
			"en-GB",
			"horse_group",
			"questions",
		},
		{
			"fibbing_it",
			uuid.New(),
			"multiple_choice",
			true,
			"What do you think about horses?",
			"en-GB",
			"horse_group",
			"questions",
		},
		{"fibbing_it", uuid.New(), "multiple_choice", true, "purple", "en-GB", "colour_group", "answers"},
		{"fibbing_it", uuid.New(), "multiple_choice", true, "blue", "en-GB", "colour_group", "answers"},
		{"fibbing_it", uuid.New(), "multiple_choice", true, "red", "en-GB", "colour_group", "answers"},
		{
			"fibbing_it",
			uuid.New(),
			"multiple_choice",
			true,
			"What is your favourite colour?",
			"en-GB",
			"colour_group",
			"questions",
		},
		{
			"fibbing_it",
			uuid.New(),
			"multiple_choice",
			true,
			"What is your least favourite colour?",
			"en-GB",
			"colour_group",
			"questions",
		},
		{"fibbing_it", uuid.New(), "multiple_choice", true, "Strongly Agree", "en-GB", "animal_group", "answers"},
		{"fibbing_it", uuid.New(), "multiple_choice", true, "Agree", "en-GB", "animal_group", "answers"},
		{"fibbing_it", uuid.New(), "multiple_choice", true, "Disagree", "en-GB", "animal_group", "answers"},
		{"fibbing_it", uuid.New(), "multiple_choice", true, "Are cats cute?", "en-GB", "animal_group", "questions"},
		{
			"fibbing_it",
			uuid.New(),
			"multiple_choice",
			true,
			"Dogs are cuter than cats?",
			"en-GB",
			"animal_group",
			"questions",
		},
	}

	for _, q := range questions {
		groupID := groupNameToID[q.GroupName][q.GroupType]

		_, err := queries.WithTx(tx).AddQuestion(ctx, db.AddQuestionParams{
			ID:        q.QuestionID,
			GameName:  q.GameName,
			RoundType: q.Round,
			GroupID:   groupID,
		})
		if err != nil {
			return err
		}

		// TODO: handle translations at the moment this code works because everythign is in en-GB.
		_, err = queries.WithTx(tx).AddQuestionTranslation(ctx, db.AddQuestionTranslationParams{
			ID:         uuid.Must(uuid.NewV7()),
			Question:   q.Question,
			Locale:     q.Locale,
			QuestionID: q.QuestionID,
		})
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}
