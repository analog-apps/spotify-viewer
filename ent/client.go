// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"
	"errors"
	"fmt"
	"log"
	"reflect"

	"github.com/albe2669/spotify-viewer/ent/migrate"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql"
	"github.com/albe2669/spotify-viewer/ent/settings"
)

// Client is the client that holds all ent builders.
type Client struct {
	config
	// Schema is the client for creating, migrating and dropping schema.
	Schema *migrate.Schema
	// Settings is the client for interacting with the Settings builders.
	Settings *SettingsClient
	// additional fields for node api
	tables tables
}

// NewClient creates a new client configured with the given options.
func NewClient(opts ...Option) *Client {
	client := &Client{config: newConfig(opts...)}
	client.init()
	return client
}

func (c *Client) init() {
	c.Schema = migrate.NewSchema(c.driver)
	c.Settings = NewSettingsClient(c.config)
}

type (
	// config is the configuration for the client and its builder.
	config struct {
		// driver used for executing database requests.
		driver dialect.Driver
		// debug enable a debug logging.
		debug bool
		// log used for logging on debug mode.
		log func(...any)
		// hooks to execute on mutations.
		hooks *hooks
		// interceptors to execute on queries.
		inters *inters
	}
	// Option function to configure the client.
	Option func(*config)
)

// newConfig creates a new config for the client.
func newConfig(opts ...Option) config {
	cfg := config{log: log.Println, hooks: &hooks{}, inters: &inters{}}
	cfg.options(opts...)
	return cfg
}

// options applies the options on the config object.
func (c *config) options(opts ...Option) {
	for _, opt := range opts {
		opt(c)
	}
	if c.debug {
		c.driver = dialect.Debug(c.driver, c.log)
	}
}

// Debug enables debug logging on the ent.Driver.
func Debug() Option {
	return func(c *config) {
		c.debug = true
	}
}

// Log sets the logging function for debug mode.
func Log(fn func(...any)) Option {
	return func(c *config) {
		c.log = fn
	}
}

// Driver configures the client driver.
func Driver(driver dialect.Driver) Option {
	return func(c *config) {
		c.driver = driver
	}
}

// Open opens a database/sql.DB specified by the driver name and
// the data source name, and returns a new client attached to it.
// Optional parameters can be added for configuring the client.
func Open(driverName, dataSourceName string, options ...Option) (*Client, error) {
	switch driverName {
	case dialect.MySQL, dialect.Postgres, dialect.SQLite:
		drv, err := sql.Open(driverName, dataSourceName)
		if err != nil {
			return nil, err
		}
		return NewClient(append(options, Driver(drv))...), nil
	default:
		return nil, fmt.Errorf("unsupported driver: %q", driverName)
	}
}

// ErrTxStarted is returned when trying to start a new transaction from a transactional client.
var ErrTxStarted = errors.New("ent: cannot start a transaction within a transaction")

// Tx returns a new transactional client. The provided context
// is used until the transaction is committed or rolled back.
func (c *Client) Tx(ctx context.Context) (*Tx, error) {
	if _, ok := c.driver.(*txDriver); ok {
		return nil, ErrTxStarted
	}
	tx, err := newTx(ctx, c.driver)
	if err != nil {
		return nil, fmt.Errorf("ent: starting a transaction: %w", err)
	}
	cfg := c.config
	cfg.driver = tx
	return &Tx{
		ctx:      ctx,
		config:   cfg,
		Settings: NewSettingsClient(cfg),
	}, nil
}

// BeginTx returns a transactional client with specified options.
func (c *Client) BeginTx(ctx context.Context, opts *sql.TxOptions) (*Tx, error) {
	if _, ok := c.driver.(*txDriver); ok {
		return nil, errors.New("ent: cannot start a transaction within a transaction")
	}
	tx, err := c.driver.(interface {
		BeginTx(context.Context, *sql.TxOptions) (dialect.Tx, error)
	}).BeginTx(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("ent: starting a transaction: %w", err)
	}
	cfg := c.config
	cfg.driver = &txDriver{tx: tx, drv: c.driver}
	return &Tx{
		ctx:      ctx,
		config:   cfg,
		Settings: NewSettingsClient(cfg),
	}, nil
}

// Debug returns a new debug-client. It's used to get verbose logging on specific operations.
//
//	client.Debug().
//		Settings.
//		Query().
//		Count(ctx)
func (c *Client) Debug() *Client {
	if c.debug {
		return c
	}
	cfg := c.config
	cfg.driver = dialect.Debug(c.driver, c.log)
	client := &Client{config: cfg}
	client.init()
	return client
}

// Close closes the database connection and prevents new queries from starting.
func (c *Client) Close() error {
	return c.driver.Close()
}

// Use adds the mutation hooks to all the entity clients.
// In order to add hooks to a specific client, call: `client.Node.Use(...)`.
func (c *Client) Use(hooks ...Hook) {
	c.Settings.Use(hooks...)
}

// Intercept adds the query interceptors to all the entity clients.
// In order to add interceptors to a specific client, call: `client.Node.Intercept(...)`.
func (c *Client) Intercept(interceptors ...Interceptor) {
	c.Settings.Intercept(interceptors...)
}

// Mutate implements the ent.Mutator interface.
func (c *Client) Mutate(ctx context.Context, m Mutation) (Value, error) {
	switch m := m.(type) {
	case *SettingsMutation:
		return c.Settings.mutate(ctx, m)
	default:
		return nil, fmt.Errorf("ent: unknown mutation type %T", m)
	}
}

// SettingsClient is a client for the Settings schema.
type SettingsClient struct {
	config
}

// NewSettingsClient returns a client for the Settings from the given config.
func NewSettingsClient(c config) *SettingsClient {
	return &SettingsClient{config: c}
}

// Use adds a list of mutation hooks to the hooks stack.
// A call to `Use(f, g, h)` equals to `settings.Hooks(f(g(h())))`.
func (c *SettingsClient) Use(hooks ...Hook) {
	c.hooks.Settings = append(c.hooks.Settings, hooks...)
}

// Intercept adds a list of query interceptors to the interceptors stack.
// A call to `Intercept(f, g, h)` equals to `settings.Intercept(f(g(h())))`.
func (c *SettingsClient) Intercept(interceptors ...Interceptor) {
	c.inters.Settings = append(c.inters.Settings, interceptors...)
}

// Create returns a builder for creating a Settings entity.
func (c *SettingsClient) Create() *SettingsCreate {
	mutation := newSettingsMutation(c.config, OpCreate)
	return &SettingsCreate{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// CreateBulk returns a builder for creating a bulk of Settings entities.
func (c *SettingsClient) CreateBulk(builders ...*SettingsCreate) *SettingsCreateBulk {
	return &SettingsCreateBulk{config: c.config, builders: builders}
}

// MapCreateBulk creates a bulk creation builder from the given slice. For each item in the slice, the function creates
// a builder and applies setFunc on it.
func (c *SettingsClient) MapCreateBulk(slice any, setFunc func(*SettingsCreate, int)) *SettingsCreateBulk {
	rv := reflect.ValueOf(slice)
	if rv.Kind() != reflect.Slice {
		return &SettingsCreateBulk{err: fmt.Errorf("calling to SettingsClient.MapCreateBulk with wrong type %T, need slice", slice)}
	}
	builders := make([]*SettingsCreate, rv.Len())
	for i := 0; i < rv.Len(); i++ {
		builders[i] = c.Create()
		setFunc(builders[i], i)
	}
	return &SettingsCreateBulk{config: c.config, builders: builders}
}

// Update returns an update builder for Settings.
func (c *SettingsClient) Update() *SettingsUpdate {
	mutation := newSettingsMutation(c.config, OpUpdate)
	return &SettingsUpdate{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// UpdateOne returns an update builder for the given entity.
func (c *SettingsClient) UpdateOne(s *Settings) *SettingsUpdateOne {
	mutation := newSettingsMutation(c.config, OpUpdateOne, withSettings(s))
	return &SettingsUpdateOne{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// UpdateOneID returns an update builder for the given id.
func (c *SettingsClient) UpdateOneID(id int) *SettingsUpdateOne {
	mutation := newSettingsMutation(c.config, OpUpdateOne, withSettingsID(id))
	return &SettingsUpdateOne{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// Delete returns a delete builder for Settings.
func (c *SettingsClient) Delete() *SettingsDelete {
	mutation := newSettingsMutation(c.config, OpDelete)
	return &SettingsDelete{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// DeleteOne returns a builder for deleting the given entity.
func (c *SettingsClient) DeleteOne(s *Settings) *SettingsDeleteOne {
	return c.DeleteOneID(s.ID)
}

// DeleteOneID returns a builder for deleting the given entity by its id.
func (c *SettingsClient) DeleteOneID(id int) *SettingsDeleteOne {
	builder := c.Delete().Where(settings.ID(id))
	builder.mutation.id = &id
	builder.mutation.op = OpDeleteOne
	return &SettingsDeleteOne{builder}
}

// Query returns a query builder for Settings.
func (c *SettingsClient) Query() *SettingsQuery {
	return &SettingsQuery{
		config: c.config,
		ctx:    &QueryContext{Type: TypeSettings},
		inters: c.Interceptors(),
	}
}

// Get returns a Settings entity by its id.
func (c *SettingsClient) Get(ctx context.Context, id int) (*Settings, error) {
	return c.Query().Where(settings.ID(id)).Only(ctx)
}

// GetX is like Get, but panics if an error occurs.
func (c *SettingsClient) GetX(ctx context.Context, id int) *Settings {
	obj, err := c.Get(ctx, id)
	if err != nil {
		panic(err)
	}
	return obj
}

// Hooks returns the client hooks.
func (c *SettingsClient) Hooks() []Hook {
	return c.hooks.Settings
}

// Interceptors returns the client interceptors.
func (c *SettingsClient) Interceptors() []Interceptor {
	return c.inters.Settings
}

func (c *SettingsClient) mutate(ctx context.Context, m *SettingsMutation) (Value, error) {
	switch m.Op() {
	case OpCreate:
		return (&SettingsCreate{config: c.config, hooks: c.Hooks(), mutation: m}).Save(ctx)
	case OpUpdate:
		return (&SettingsUpdate{config: c.config, hooks: c.Hooks(), mutation: m}).Save(ctx)
	case OpUpdateOne:
		return (&SettingsUpdateOne{config: c.config, hooks: c.Hooks(), mutation: m}).Save(ctx)
	case OpDelete, OpDeleteOne:
		return (&SettingsDelete{config: c.config, hooks: c.Hooks(), mutation: m}).Exec(ctx)
	default:
		return nil, fmt.Errorf("ent: unknown Settings mutation op: %q", m.Op())
	}
}

// hooks and interceptors per client, for fast access.
type (
	hooks struct {
		Settings []ent.Hook
	}
	inters struct {
		Settings []ent.Interceptor
	}
)
