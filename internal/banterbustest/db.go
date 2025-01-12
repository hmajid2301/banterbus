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
	"github.com/mdobak/go-xerrors"
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
		return pool, fmt.Errorf("failed to get database :%w", err)
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
		return pool, fmt.Errorf("failed to parse db uri :%w", err)
	}

	pgxConfig.AfterConnect = func(_ context.Context, conn *pgx.Conn) error {
		pgxUUID.Register(conn.TypeMap())
		return nil
	}
	pool, err = pgxpool.NewWithConfig(ctx, pgxConfig)
	if err != nil {
		return pool, fmt.Errorf("failed to setup database :%w", err)
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
		return fmt.Errorf("failed to connect to database :%w", err)
	}
	defer pool.Close()

	_, err = pool.Exec(ctx, fmt.Sprintf("DROP DATABASE IF EXISTS %s", dbName))
	if err != nil {
		return fmt.Errorf("failed to drop database :%w", err)
	}

	return nil
}

func getDatabaseName(pool *pgxpool.Pool) (string, error) {
	connConfig := pool.Config().ConnConfig
	connString := connConfig.ConnString()

	config, err := pgx.ParseConfig(connString)
	if err != nil {
		return "", fmt.Errorf("failed to parse connection string :%w", err)
	}

	return config.Database, nil
}

func runDBMigrations(pool *pgxpool.Pool) error {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return xerrors.New("failed to get current filename")
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

type Group struct {
	ID   string
	Name string
	Type string
}

func FillWithDummyData(ctx context.Context, pool *pgxpool.Pool) error {
	tx, err := pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}

	queries := db.New(pool)
	groups := []Group{
		{ID: "01945c66-891a-7894-ae92-c18087c73a23", Name: "programming_group", Type: "questions"},
		{ID: "01945c66-891b-797e-820b-17c4cc98a566", Name: "programming_group", Type: "answers"},
		{ID: "01945c66-891c-7942-9a2a-339a62a74800", Name: "horse_group", Type: "questions"},
		{ID: "01945c66-891c-7794-ab47-8f70974b0037", Name: "horse_group", Type: "answers"},
		{ID: "01945c66-891c-7aa2-b6ca-088679706a5b", Name: "colour_group", Type: "questions"},
		{ID: "01945c66-891c-7be5-9382-7b6399a6b09b", Name: "colour_group", Type: "answers"},
		{ID: "01945c66-891b-7d3e-804c-f2e170b0b0ce", Name: "cat_group", Type: "questions"},
		{ID: "01945c66-891b-7bb1-875f-a2be41d430ab", Name: "cat_group", Type: "answers"},
		{ID: "01945c66-891c-74d5-9870-7a8777e37588", Name: "bike_group", Type: "questions"},
		{ID: "01945c66-891c-717f-a177-3240919b638f", Name: "bike_group", Type: "answers"},
		{ID: "01945c66-891c-7d8a-b404-be384c9515a6", Name: "animal_group", Type: "questions"},
		{ID: "01945c66-891c-7ee0-88fb-5703de2a1b59", Name: "animal_group", Type: "answers"},
		{ID: "01945c66-891d-723b-8fc9-b373d799e95b", Name: "all", Type: "questions"},
		{ID: "01945c66-891d-70fc-945a-e30fa7b09bfb", Name: "all", Type: "answers"},
	}

	groupNameToID := map[string]map[string]uuid.UUID{}

	for _, group := range groups {
		questionGroup, err := queries.WithTx(tx).AddGroup(ctx, db.AddGroupParams{
			ID:        uuid.MustParse(group.ID),
			GroupName: group.Name,
			GroupType: group.Type,
		})
		if err != nil {
			return err
		}
		if _, ok := groupNameToID[group.Name]; !ok {
			groupNameToID[group.Name] = map[string]uuid.UUID{}
		}

		groupNameToID[group.Name][group.Type] = questionGroup.ID
	}

	questions := []struct {
		GameName   string
		QuestionID string
		Round      string
		Enabled    bool
		Question   string
		Locale     string
		GroupName  string
		GroupType  string
	}{
		{
			"fibbing_it",
			"4b1355bb-82de-40c8-8eda-0c634091cc3c",
			"most_likely",
			false,
			"to get arrested",
			"en-GB",
			"all",
			"questions",
		},
		{
			"fibbing_it",
			"a91af98c-f989-4e00-aa14-7a34e732519e",
			"most_likely",
			true,
			"to eat ice-cream from the tub",
			"en-GB",
			"all",
			"questions",
		},
		{
			"fibbing_it",
			"fac6a98f-e3b5-4328-999c-b39fd86657ba",
			"most_likely",
			true,
			"to fight a police person",
			"en-GB",
			"all",
			"questions",
		},
		{
			"fibbing_it",
			"6b60f097-b714-4f9e-b8cb-de75a7890381",
			"most_likely",
			true,
			"to fight a horse",
			"en-GB",
			"all",
			"questions",
		},
		{
			"fibbing_it",
			"93dd56a8-c8a3-4c63-93dc-9d890c4d2b74",
			"free_form",
			true,
			"What do you think about programmers?",
			"en-GB",
			"programming_group",
			"questions",
		},
		{
			"fibbing_it",
			"066e7a8a-b0b7-44d4-b882-582a64151c15",
			"free_form",
			true,
			"What don't you like about programmers?",
			"en-GB",
			"programming_group",
			"questions",
		},
		{
			"fibbing_it",
			"654327b9-36a2-4d75-b4bf-d68d19fcfe7c",
			"free_form",
			true,
			"what don't you think about programmers?",
			"en-GB",
			"programming_group",
			"questions",
		},
		{
			"fibbing_it",
			"281bc3c7-f55d-4a8a-88cf-4e0d67d2825e",
			"free_form",
			true,
			"what dont you think about cats",
			"en-GB",
			"cat_group",
			"questions",
		},
		{
			"fibbing_it",
			"fc1a3c9f-3d98-452e-b77e-c6c7f353176d",
			"free_form",
			true,
			"what don't you like about cats?",
			"en-GB",
			"cat_group",
			"questions",
		},
		{
			"fibbing_it",
			"393dae17-84fe-449d-ba0f-8c9d320a46e6",
			"free_form",
			false,
			"what do you like about cats?",
			"en-GB",
			"cat_group",
			"questions",
		},
		{
			"fibbing_it",
			"393dae17-84fe-449d-ba0f-8c9d320a46e7",
			"free_form",
			true,
			"what do you think about cats",
			"en-GB",
			"cat_group",
			"questions",
		},
		{
			"fibbing_it",
			"9347fb72-77f9-44c4-9a7c-27109e29dd97",
			"free_form",
			true,
			"A funny question?",
			"en-GB",
			"bike_group",
			"questions",
		},
		{
			"fibbing_it",
			"8aa9f87f-31d9-4421-aae5-2024ca730348",
			"free_form",
			true,
			"Favourite bike colour?",
			"en-GB",
			"bike_group",
			"questions",
		},
		{
			"fibbing_it",
			"89b20c84-12ae-444d-ad9c-26f72d3f28ab",
			"multiple_choice",
			true,
			"What do you think about camels?",
			"en-GB",
			"horse_group",
			"questions",
		},
		{
			"fibbing_it",
			"68ed9133-dc58-41bb-b642-c48470998127",
			"multiple_choice",
			true,
			"What do you think about horses?",
			"en-GB",
			"horse_group",
			"questions",
		},
		{
			"fibbing_it",
			"56dae1b2-06e9-4339-a2b0-892a18444e15",
			"multiple_choice",
			true,
			"What is your favourite colour?",
			"en-GB",
			"colour_group",
			"questions",
		},
		{
			"fibbing_it",
			"8a76fbb3-c9ad-47b2-a195-9d8623ab8da0",
			"multiple_choice",
			true,
			"What is your least favourite colour?",
			"en-GB",
			"colour_group",
			"questions",
		},
		{
			"fibbing_it",
			"cb9b99be-e66f-4f1b-9a62-64415a824b31",
			"multiple_choice",
			true,
			"Strongly Agree",
			"en-GB",
			"animal_group",
			"answers",
		},
		{
			"fibbing_it",
			"e72b7503-88ea-440e-a42c-bc6f2f444b08",
			"multiple_choice",
			true,
			"Agree",
			"en-GB",
			"animal_group",
			"answers",
		},
		{
			"fibbing_it",
			"fa459292-bca0-47ec-81f1-5fb48036f3ea",
			"multiple_choice",
			true,
			"Disagree",
			"en-GB",
			"animal_group",
			"answers",
		},
		{
			"fibbing_it",
			"e90d613d-2e6c-4331-9204-9b685c0795b7",
			"multiple_choice",
			true,
			"Are cats cute?",
			"en-GB",
			"animal_group",
			"questions",
		},
		{
			"fibbing_it",
			"89deb03f-66be-4265-91e6-dedd9227718a",
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
			ID:        uuid.MustParse(q.QuestionID),
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
			QuestionID: uuid.MustParse(q.QuestionID),
		})
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}
