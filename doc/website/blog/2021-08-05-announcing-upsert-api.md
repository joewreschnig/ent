---
title: Announcing the Upsert API in v0.9.0
author: Rotem Tamir
authorURL: "https://github.com/rotemtam"
authorImageURL: "https://s.gravatar.com/avatar/36b3739951a27d2e37251867b7d44b1a?s=80"
authorTwitter: _rtam
---


It has been almost 4 months since our [last release](https://github.com/ent/ent/releases/tag/v0.8.0), and for a good reason.
Version [0.9.0](https://github.com/ent/ent/releases/tag/v0.9.0) which was released today is packed with some highly-anticipated
features. Perhaps at the top of the list, is a feature that has been in discussion for [more than a year in a half](https://github.com/ent/ent/issues/139)
and was one of the most commonly requested features in the [Ent User Survey](https://forms.gle/7VZSPVc7D1iu75GV9): the Upsert API!

Version 0.9.0 adds support for "Upsert" style statements using [a new feature flag](https://entgo.io/docs/feature-flags#upsert): `sql/upsert`. 
Ent has a [collection of feature flags](https://entgo.io/docs/feature-flags) that can be switched on to add more features
to the code generated by Ent. This is used as both a mechanism to allow opt-in to some features that are not necessarily
desired in every project and as a way to run experiments of features that may one day become part of Ent's core.

In this post, we will introduce the new feature, the places where it is useful, and demonstrate how to use it.

### Upsert

"Upsert" is a commonly-used term in data systems that is a portmanteau of "update" and "insert" which usually refers to
a statement that attempts to insert a record to a table, and if a uniqueness constraint is violated (e.g. a record by
that ID already exists) that record is updated instead. While none of the popular relational databases have a specific
`UPSERT` statement, most of them support ways of achieving this type of behavior.

For example, assume we have a table with this definition in an SQLite database:

```sql
CREATE TABLE users (
   id integer PRIMARY KEY AUTOINCREMENT,
   email varchar(255) UNIQUE,
   name varchar(255)
)
```

If we try to execute the same insert twice:

```sql
INSERT INTO users (email, name) VALUES ('rotem@entgo.io', 'Rotem Tamir');
INSERT INTO users (email, name) VALUES ('rotem@entgo.io', 'Rotem Tamir');
```

We get this error:

```
[2021-08-05 06:49:22] UNIQUE constraint failed: users.email
```

In many cases, it is useful to have write operations be [idempotent](https://en.wikipedia.org/wiki/Idempotence),
meaning we can run them many times in a row while leaving the system in the same state.

In other cases, it is not desirable to query if a record exists before trying to create it. For these kinds of situations,
SQLite supports the [`ON CONFLICT` clause](https://www.sqlite.org/lang_upsert.html) in `INSERT`
statements. To instruct SQLite to override an existing value with the new one we can execute:

```sql
INSERT INTO users (email, name) values ('rotem@entgo.io', 'Tamir, Rotem')
ON CONFLICT (email) DO UPDATE SET email=excluded.email, name=excluded.name;
```

If we prefer to keep the existing values, we can use the `DO NOTHING` conflict action:

```sql
INSERT INTO users (email, name) values ('rotem@entgo.io', 'Tamir, Rotem') 
ON CONFLICT DO NOTHING;
```

Sometimes we want to merge the two versions in some way, we can use the `DO UPDATE` action a little differently to
achieve do something like:

```sql
INSERT INTO users (email, full_name) values ('rotem@entgo.io', 'Tamir, Rotem') 
ON CONFLICT (email) DO UPDATE SET name=excluded.name ||  ' (formerly: ' || users.name || ')'
```

In this case, after our second `INSERT` the value for the `name` column would be: `Tamir, Rotem (formerly: Rotem Tamir)`.
Not very useful, but hopefully you can see that you can do cool things this way.

### Upsert with Ent

Assume we have an existing Ent project with an entity similar to the `users` table described above:

```go
// User holds the schema definition for the User entity.
type User struct {
	ent.Schema
}

// Fields of the User.
func (User) Fields() []ent.Field {
	return []ent.Field{
		field.String("email").
			Unique(),
		field.String("name"),
	}
}
```

As the Upsert API is a newly released feature, make sure to update your `ent` version using:

```bash
go get -u entgo.io/ent@v0.9.0
```

Next, add the `sql/upsert` feature flag to your code-generation flags, in `ent/generate.go`:

```go
package ent

//go:generate go run -mod=mod entgo.io/ent/cmd/ent generate --feature sql/upsert ./schema
```

Next, re-run code generation for your project:

```go
go generate ./...
```

Observe that a new method named `OnConflict` was added to the `ent/user_create.go` file:

```go
// OnConflict allows configuring the `ON CONFLICT` / `ON DUPLICATE KEY` clause
// of the `INSERT` statement. For example:
//
//	client.User.Create().
//		SetEmailAddress(v).
//		OnConflict(
//			// Update the row with the new values
//			// the was proposed for insertion.
//			sql.ResolveWithNewValues(),
//		).
//		// Override some of the fields with custom
//		// update values.
//		Update(func(u *ent.UserUpsert) {
//			SetEmailAddress(v+v)
//		}).
//		Exec(ctx)
//
func (uc *UserCreate) OnConflict(opts ...sql.ConflictOption) *UserUpsertOne {
	uc.conflict = opts
	return &UserUpsertOne{
		create: uc,
	}
}
```

This (along with more new generated code) will serve us in achieving upsert behavior for our `User` entity.
To explore this, let's first start by writing a test to reproduce the uniqueness constraint error:

```go
func TestUniqueConstraintFails(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")
	ctx := context.TODO()

	// Create the user for the first time.
	client.User.
		Create().
		SetEmail("rotem@entgo.io").
		SetName("Rotem Tamir").
		SaveX(ctx)

	// Try to create a user with the same email the second time.
	_, err := client.User.
		Create().
		SetEmail("rotem@entgo.io").
		SetName("Rotem Tamir").
		Save(ctx)

	if !ent.IsConstraintError(err) {
		log.Fatalf("expected second created to fail with constraint error")
	}
	log.Printf("second query failed with: %v", err)
}
```

The test passes:

```bash
=== RUN   TestUniqueConstraintFails
2021/08/05 07:12:11 second query failed with: ent: constraint failed: insert node to table "users": UNIQUE constraint failed: users.email
--- PASS: TestUniqueConstraintFails (0.00s)
```

Next, let's see how to instruct Ent to override the existing values with the new in case a conflict occurs:

```go
func TestUpsertReplace(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")
	ctx := context.TODO()

	// Create the user for the first time.
	orig := client.User.
		Create().
		SetEmail("rotem@entgo.io").
		SetName("Rotem Tamir").
		SaveX(ctx)

	// Try to create a user with the same email the second time.
	// This time we set ON CONFLICT behavior, and use the `UpdateNewValues`
	// modifier.
	newID := client.User.Create().
		SetEmail("rotem@entgo.io").
		SetName("Tamir, Rotem").
		OnConflict().
		UpdateNewValues().
		// we use the IDX method to receive the ID
		// of the created/updated entity
		IDX(ctx)

	// We expect the ID of the originally created user to be the same as
	// the one that was just updated.
	if orig.ID != newID {
		log.Fatalf("expected upsert to update an existing record")
	}

	current := client.User.GetX(ctx, orig.ID)
	if current.Name != "Tamir, Rotem" {
		log.Fatalf("expected upsert to replace with the new values")
	}
}
```

Running our test:

```bash
=== RUN   TestUpsertReplace
--- PASS: TestUpsertReplace (0.00s)
```

Alternatively, we can use the `Ignore` modifier to instruct Ent to keep the old version when resolving the conflict.
Let's write a test that shows this:

```go
func TestUpsertIgnore(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")
	ctx := context.TODO()

	// Create the user for the first time.
	orig := client.User.
		Create().
		SetEmail("rotem@entgo.io").
		SetName("Rotem Tamir").
		SaveX(ctx)

	// Try to create a user with the same email the second time.
	// This time we set ON CONFLICT behavior, and use the `Ignore`
	// modifier.
	client.User.
		Create().
		SetEmail("rotem@entgo.io").
		SetName("Tamir, Rotem").
		OnConflict().
		Ignore().
		ExecX(ctx)

	current := client.User.GetX(ctx, orig.ID)
	if current.FullName != orig.FullName {
		log.Fatalf("expected upsert to keep the original version")
	}
}
```

You can read more about the feature in the [Feature Flag](https://entgo.io/docs/feature-flags#upsert) or [Upsert API](https://entgo.io/docs/crud#upsert-one) documentation.

### Wrapping Up

In this post, we presented the Upsert API, a long-anticipated capability, that is available by feature-flag in Ent v0.9.0.
We discussed where upserts are commonly used in applications and the way they are implemented using common relational databases.
Finally, we showed a simple example of how to get started with the Upsert API using Ent.

Have questions? Need help with getting started? Feel free to [join our Slack channel](https://entgo.io/docs/slack/).

:::note For more Ent news and updates:

- Subscribe to our [Newsletter](https://www.getrevue.co/profile/ent)
- Follow us on [Twitter](https://twitter.com/entgo_io)
- Join us on #ent on the [Gophers Slack](https://entgo.io/docs/slack)

:::