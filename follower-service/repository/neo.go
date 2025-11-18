package repository

import (
    "context"
    "fmt"

    "github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

type NeoRepository struct {
    driver neo4j.DriverWithContext
}

func NewNeoRepository(ctx context.Context, uri, user, password string) (*NeoRepository, error) {
    auth := neo4j.BasicAuth(user, password, "")
    driver, err := neo4j.NewDriverWithContext(uri, auth)
    if err != nil {
        return nil, err
    }
    if err != nil {
        driver.Close(ctx)
        return nil, fmt.Errorf("verify session: %w", err)
    }
    return &NeoRepository{driver: driver}, nil
}

func (r *NeoRepository) Close() error {
    return r.driver.Close(context.Background())
}

func (r *NeoRepository) Follow(ctx context.Context, follower, followee string) error {
    sess := r.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
    defer sess.Close(ctx)
    _, err := sess.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
        cy := `MERGE (a:User {id:$follower}) MERGE (b:User {id:$followee}) MERGE (a)-[:FOLLOWS]->(b)`
        params := map[string]any{"follower": follower, "followee": followee}
        _, err := tx.Run(ctx, cy, params)
        return nil, err
    })
    return err
}

func (r *NeoRepository) Unfollow(ctx context.Context, follower, followee string) (int64, error) {
    sess := r.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
    defer sess.Close(ctx)
    res, err := sess.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
        cy := `MATCH (a:User {id:$follower})-[rel:FOLLOWS]->(b:User {id:$followee}) DELETE rel RETURN count(rel) as deleted` 
        params := map[string]any{"follower": follower, "followee": followee}
        rec, err := tx.Run(ctx, cy, params)
        if err != nil {
            return int64(0), err
        }
        var deleted int64 = 0
        if rec.Next(ctx) {
            if v, ok := rec.Record().Values[0].(int64); ok {
                deleted = v
            }
        }
        return deleted, rec.Err()
    })
    if err != nil {
        return 0, err
    }
    return res.(int64), nil
}

func (r *NeoRepository) Followers(ctx context.Context, userID string) ([]string, error) {
    sess := r.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
    defer sess.Close(ctx)
    res, err := sess.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
        cy := `MATCH (f:User)-[:FOLLOWS]->(u:User {id:$id}) RETURN f.id as id` 
        params := map[string]any{"id": userID}
        rec, err := tx.Run(ctx, cy, params)
        if err != nil {
            return nil, err
        }
        var out []string
        for rec.Next(ctx) {
            recMap := rec.Record().Values
            if len(recMap) > 0 {
                if v, ok := recMap[0].(string); ok {
                    out = append(out, v)
                }
            }
        }
        return out, rec.Err()
    })
    if err != nil {
        return nil, err
    }
    return res.([]string), nil
}

func (r *NeoRepository) Following(ctx context.Context, userID string) ([]string, error) {
    sess := r.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
    defer sess.Close(ctx)
    res, err := sess.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
        cy := `MATCH (u:User {id:$id})-[:FOLLOWS]->(f:User) RETURN f.id as id` 
        params := map[string]any{"id": userID}
        rec, err := tx.Run(ctx, cy, params)
        if err != nil {
            return nil, err
        }
        var out []string
        for rec.Next(ctx) {
            recMap := rec.Record().Values
            if len(recMap) > 0 {
                if v, ok := recMap[0].(string); ok {
                    out = append(out, v)
                }
            }
        }
        return out, rec.Err()
    })
    if err != nil {
        return nil, err
    }
    return res.([]string), nil
}

func (r *NeoRepository) Recommendations(ctx context.Context, userID string, limit int) ([]string, error) {
    sess := r.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
    defer sess.Close(ctx)
    res, err := sess.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
        cy := `MATCH (u:User {id:$id})-[:FOLLOWS]->(f)-[:FOLLOWS]->(rec:User)
WHERE NOT (u)-[:FOLLOWS]->(rec) AND rec.id <> $id
RETURN rec.id as id, count(*) as cnt
ORDER BY cnt DESC
LIMIT $limit`
        params := map[string]any{"id": userID, "limit": limit}
        rec, err := tx.Run(ctx, cy, params)
        if err != nil {
            return nil, err
        }
        var out []string
        for rec.Next(ctx) {
            vals := rec.Record().Values
            if len(vals) > 0 {
                if v, ok := vals[0].(string); ok {
                    out = append(out, v)
                }
            }
        }
        return out, rec.Err()
    })
    if err != nil {
        return nil, err
    }
    return res.([]string), nil
}
