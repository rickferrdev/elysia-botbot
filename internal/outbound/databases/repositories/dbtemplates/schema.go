package dbtemplates

import (
	"time"

	"github.com/rickferrdev/elysia-botbot/internal/core/domain"
	"github.com/uptrace/bun"
)

type SchemaTemplate struct {
	bun.BaseModel `bun:"table:template,alias:t"`

	ID       int    `bun:"id,pk,autoincrement"`
	GuildID  string `bun:"guild_id,notnull"`
	AuthorID string `bun:"author_id,notnull"`

	CreatedAt time.Time `bun:"created_at"`
	UpdatedAt time.Time `bun:"updated_at"`

	TemplateContent *SchemaTemplateContent `bun:"rel:has-one,join:id=template_id"`
}

type SchemaTemplateContent struct {
	bun.BaseModel `bun:"table:template_contents"`

	ID         int `bun:"id,pk,autoincrement"`
	TemplateID int `bun:"template_id,notnull"`

	Content     string `bun:"content"`
	Attachments string `bun:"attachments"`
	Embeds      string `bun:"embeds"`

	CreatedAt time.Time `bun:"created_at"`
	UpdatedAt time.Time `bun:"updated_at"`

	Template *SchemaTemplate `bun:"rel:belongs-to,join:template_id=id"`
}

func NewSchemaTemplate(template domain.Template) SchemaTemplate {
	schema := SchemaTemplate{
		ID:        template.ID,
		GuildID:   template.GuildID,
		AuthorID:  template.AuthorID,
		CreatedAt: template.CreatedAt,
		UpdatedAt: template.UpdatedAt,
	}

	return schema
}

func (schema *SchemaTemplate) NewDomainTemplate() *domain.Template {
	domain := domain.Template{
		ID:        schema.ID,
		GuildID:   schema.GuildID,
		AuthorID:  schema.AuthorID,
		CreatedAt: schema.CreatedAt,
		UpdatedAt: schema.UpdatedAt,
	}
	if schema.TemplateContent != nil {
		content := schema.TemplateContent.NewDomainTemplateContent()
		domain.TemplateContent = &content
	}

	return &domain
}

func NewSchemaTemplateContent(template domain.TemplateContent) SchemaTemplateContent {
	schema := SchemaTemplateContent{
		ID:          template.ID,
		TemplateID:  template.TemplateID,
		Attachments: template.Attachments,
		Content:     template.Content,
		Embeds:      template.Embeds,
		CreatedAt:   template.CreatedAt,
		UpdatedAt:   template.UpdatedAt,
	}

	return schema
}

func (schema *SchemaTemplateContent) NewDomainTemplateContent() domain.TemplateContent {
	domain := domain.TemplateContent{
		ID:          schema.ID,
		TemplateID:  schema.TemplateID,
		Attachments: schema.Attachments,
		Content:     schema.Content,
		Embeds:      schema.Embeds,
		CreatedAt:   schema.CreatedAt,
		UpdatedAt:   schema.UpdatedAt,
	}

	return domain
}
