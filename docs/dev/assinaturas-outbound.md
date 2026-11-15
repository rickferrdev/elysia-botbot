# Assinaturas do outbound

## Repositórios

### `dbautoreactions`

```go
func New(args FxArgs) (*Database, error)

func (database *Database) CreateOne(ctx context.Context, autoreactions domain.AutoReactions) (*domain.AutoReactions, error)
func (database *Database) GetOne(ctx context.Context, guildID, channelID string) (*domain.AutoReactions, error)
func (database *Database) GetAll(ctx context.Context, guildID string) ([]domain.AutoReactions, error)
func (database *Database) UpdateOne(ctx context.Context, autoreactions domain.AutoReactions) (*domain.AutoReactions, error)
func (database *Database) DeleteOne(ctx context.Context, guildID, channelID string) (*domain.AutoReactions, error)
```

### `dbservers`

```go
func New(args FxArgs) (*Database, error)

func (database *Database) CreateServerWithConfig(ctx context.Context, server domain.Server) (*domain.Server, error)
func (database *Database) GetServerWithConfig(ctx context.Context, guildID string) (*domain.Server, error)
```

### `dbslipper`
- deprecated

```go
func New(args FxArgs) (*Database, error)

func (database *Database) IncrementWith(ctx context.Context, guildID, userID string, value int) (int, error)
func (database *Database) DecrementWith(ctx context.Context, guildID, userID string, value int) (int, error)
func (database *Database) GetInventory(ctx context.Context, guildID, userID string) (int, error)
```

### `dbtomatoes`
- deprecated

```go
func New(args FxArgs) (*Database, error)

func (database *Database) Increment(ctx context.Context, guildID, userID string, value int) (int, error)
func (database *Database) Decrement(ctx context.Context, guildID, userID string, value int) (int, error)
func (database *Database) GetInventory(ctx context.Context, guildID, userID string) (int, error)
```

### `dbtemplates`

```go
func New(args FxArgs) (*Database, error)

func (database *Database) CreateOne(ctx context.Context, template domain.Template) (*domain.Template, error)
func (database *Database) GetOne(ctx context.Context, id int) (*domain.Template, error)
func (database *Database) GetAll(ctx context.Context, guildID string, authorID string) ([]domain.Template, error)
func (database *Database) UpdateOne(ctx context.Context, template domain.Template) (*domain.Template, error)
func (database *Database) DeleteOne(ctx context.Context, id int) (*domain.Template, error)
```

### `dbstore`

```go
type CatalogInterface interface {
	FindAll(ctx context.Context, limit int) (*[]domain.Catalog, error)
	FindByName(ctx context.Context, name string) (*domain.Catalog, error)
	Insert(ctx context.Context, catalog domain.Catalog) error
}

func NewCatalog(args CatalogFxArgs) (*CatalogDatabase, error)

func (database *CatalogDatabase) FindAll(ctx context.Context, limit int) (*[]domain.Catalog, error)
func (database *CatalogDatabase) FindByName(ctx context.Context, name string) (*domain.Catalog, error)
```

## Testes

### `dbautoreactions`

```go
func (test *test) TestCreateOne()
func (test *test) TestCreateUpsert()
func (test *test) TestGetOne()
func (test *test) TestGetOneNotFound()
func (test *test) TestGetAll()
func (test *test) TestGetAllNotFound()
func (test *test) TestUpdateOne()
func (test *test) TestUpdateNotFound()
func (test *test) TestDeleteOne()
func (test *test) TestDeleteOneNotFound()
func TestRunner(t *testing.T)
```

### `dbservers`

```go
func (test *test) TestCreateServerWithConfig()
func (test *test) TestCreateServerWithInvalidConfig()
func (test *test) TestGetServerWithConfig()
func (test *test) TestGetServerWithConfigNotFound()
func TestRunner(t *testing.T)
```

### `dbslipper`

```go
func (test *test) TestIncrementWith()
func (test *test) TestDecrement()
func (test *test) TestGetInventory()
func (test *test) TestGetInventoryNotFound()
func TestRunner(t *testing.T)
```

### `dbtomatoes`

```go
func (test *test) TestIncrement()
func (test *test) TestDecrement()
func (test *test) TestGetInventory()
func (test *test) TestGetInventoryNotFound()
func TestRunner(t *testing.T)
```

### `dbtemplates`

```go
func (test *test) TestCreateOne()
func (test *test) TestGetOne()
func (test *test) TestGetOneNotFound()
func (test *test) TestGetAll()
func (test *test) TestGetAllNotFound()
func (test *test) TestUpdateOne()
func (test *test) TestUpdateOneNotFound()
func (test *test) TestDeleteOne()
func (test *test) TestDeleteOneNotFound()
func TestRunner(t *testing.T)
```

### `dbstore`

```go
func (test *test) TestFindAll()
func (test *test) TestFindByName()
func (test *test) TestFindByNameNotFound()
func (test *test) TestFindAllNotFound()
func TestRunner(t *testing.T)
```
