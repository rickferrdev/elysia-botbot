package resembeds

import (
	"fmt"
	"math/rand/v2"
	"strings"

	"github.com/discord-go/discord.go/messages"
	"github.com/rickferrdev/elysia-botbot/internal/core/domain"
)

func EmbedInsufficientBalance(thumbnail string) messages.Embed {
	return messages.NewEmbedBuilder().
		SetTitle("⋆. 𐙚˚࿔ Saldo Insuficiente 𝜗𝜚˚⋆").
		SetDescription(fmt.Sprintf("Você está sem tomates 🍅 para jogar. Cada tomate custa %d chinelos 🩴.", domain.TomatoesCost)).
		SetThumbnail(thumbnail).
		SetColor(0xF38BA8).
		SetTimestampNow().
		Build()
}

func EmbedStartupShop(thumbnail string) messages.Embed {
	return messages.NewEmbedBuilder().
		SetTitle("⋆. 𐙚˚࿔ Lojinha da Elysia 𝜗𝜚˚⋆").
		SetDescription("Seja bem-vindo(a) à nossa lojinha! Escolha um item no menu abaixo para realizar a compra. ♡").
		SetThumbnail(thumbnail).
		SetColor(0xB4E4D6).
		AddField(
			"🍅 Tomates",
			fmt.Sprintf("── .✦ Preço: `%d` chinelos cada\n⋆˚꩜｡ Use para arremessar nos outros com `/tomato`!", domain.TomatoesCost),
			false,
		).
		SetTimestampNow().
		Build()
}

func EmbedPurchaseShop(thumbnail string, quantity int, cost int, balance int) messages.Embed {
	embed := messages.NewEmbedBuilder().
		SetTitle("⋆. 𐙚˚࿔ Compra Concluída! 𝜗𝜚˚⋆").
		SetDescription("Sua compra foi realizada com sucesso! Aproveite seus tomates\n🍅 Tomates. ♡").
		SetThumbnail(thumbnail).
		SetColor(0xB4E4D6).
		SetTimestampNow()

	embed.AddField(
		"── .✦ Quantidade comprada:",
		fmt.Sprintf("%dx 🍅", quantity),
		false,
	)

	embed.AddField(
		"⋆˚꩜｡ Custo total:",
		fmt.Sprintf("%dx 🩴", cost),
		false,
	)

	embed.AddField(
		"꩜｡ Novo saldo de tomates:",
		fmt.Sprintf("%dx 🍅", balance),
		false,
	)

	return embed.Build()
}

func EmbedInvalidConfigServer(thumbnail string) messages.Embed {
	return messages.NewEmbedBuilder().
		SetTitle("⋆. 𐙚˚࿔ Nenhuma configuração encontrada 𝜗𝜚˚⋆").
		SetDescription("Experimente criar uma nova para seu servidor. ♡").
		SetThumbnail(thumbnail).
		SetColor(0xB4E4D6).
		SetTimestampNow().
		Build()
}

func EmbedConfigurationsServer(thumbnail string, data domain.Server) messages.Embed {
	embed := messages.NewEmbedBuilder().
		SetTitle("⋆. 𐙚˚࿔ Configurações do Servidor 𝜗𝜚˚⋆").
		SetDescription("Aqui estão as configurações registradas para seu servidor. ♡").
		SetThumbnail(thumbnail).
		SetColor(0xB4E4D6).
		SetTimestampNow()

	embed.AddField(
		"⚙️ Servidor",
		fmt.Sprintf("── .✦ Identificador: **%v**", data.GuildID),
		false,
	)

	embed.AddField(
		"⋆˚꩜｡ Prefixo",
		fmt.Sprintf("`%v`", data.Config.Prefix),
		false,
	)

	return embed.Build()
}

func EmbedCreateEmbeds(thumbnail string, data domain.Template, source string) messages.Embed {
	embed := messages.NewEmbedBuilder().
		SetTitle("⋆. 𐙚˚࿔ Criação de Incorporação 𝜗𝜚˚⋆").
		SetDescription("Sua incorporação foi criada com sucesso! ♡").
		SetThumbnail(thumbnail).
		SetColor(0xB4E4D6).
		SetTimestampNow()

	embed.AddField(
		"📋 Identificador",
		fmt.Sprintf("`%v`", data.ID),
		false,
	)

	embed.AddField(
		"⋆˚꩜｡ Tipo de envio",
		fmt.Sprintf("`%v`", source),
		false,
	)

	embed.AddField(
		"⋆˚꩜｡ Autor",
		fmt.Sprintf("<@%v>", data.AuthorID),
		false,
	)

	embed.AddField(
		"⋆˚꩜｡ Servidor",
		fmt.Sprintf("**%v**", data.GuildID),
		false,
	)

	embed.AddField(
		"⋆˚꩜｡ Status",
		"`Criada com sucesso!`",
		false,
	)

	return embed.Build()
}

func EmbedObtainEmbeds(thumbnail string, data domain.Template, decode domain.EmbedContentDecode) messages.Embed {
	embed := messages.NewEmbedBuilder().
		SetTitle("⋆. 𐙚˚࿔ Obter uma Incorporação 𝜗𝜚˚⋆").
		SetDescription("Aqui estão os dados da incorporação consultada. ♡").
		SetThumbnail(thumbnail).
		SetColor(0xB4E4D6).
		SetTimestampNow()

	embed.AddField(
		"📄 Identificador",
		fmt.Sprintf("`%v`", data.ID),
		false,
	)

	embed.AddField(
		"⋆˚꩜｡ Criado por",
		fmt.Sprintf("<@%v>", data.AuthorID),
		false,
	)

	embed.AddField(
		"⋆˚꩜｡ Servidor",
		fmt.Sprintf("**%v**", data.GuildID),
		false,
	)

	embed.AddField(
		"⋆˚꩜｡ Total de embeds",
		fmt.Sprintf("`%d`", len(decode.Embeds)),
		false,
	)

	if strings.TrimSpace(decode.Content) != "" {
		contentPreview := decode.Content

		if len(contentPreview) > 80 {
			contentPreview = contentPreview[:77] + "..."
		}

		embed.AddField(
			"⋆˚꩜｡ Conteúdo de texto",
			fmt.Sprintf("`%v`", contentPreview),
			false,
		)
	}

	return embed.Build()
}

func EmbedObtainAllEmbeds(thumbnail string, data []domain.Template) messages.Embed {
	embed := messages.NewEmbedBuilder().
		SetTitle("⋆. 𐙚˚࿔ Obter uma Incorporação 𝜗𝜚˚⋆").
		SetDescription("Lista de todas as incorporações registradas para este servidor. ♡").
		SetThumbnail(thumbnail).
		SetColor(0xB4E4D6).
		SetTimestampNow()

	for _, value := range data {
		embed.AddField(
			fmt.Sprintf("📄 Incorporação #%v", value.ID),
			fmt.Sprintf("── .✦ Criado por: <@%v>", value.AuthorID),
			false,
		)

		embed.AddField(
			"⋆˚꩜｡ Servidor",
			fmt.Sprintf("**%v**", value.GuildID),
			false,
		)
	}

	return embed.Build()
}

func EmbedPublishEmbeds(thumbnail string, channelID string, postAuthor string, data domain.Template) messages.Embed {
	embed := messages.NewEmbedBuilder().
		SetTitle("⋆. 𐙚˚࿔ Publicar uma Incorporação 𝜗𝜚˚⋆").
		SetDescription("Sua incorporação foi publicada no canal selecionado! ♡").
		SetThumbnail(thumbnail).
		SetColor(0xB4E4D6).
		SetTimestampNow()

	embed.AddField(
		"📢 Identificador",
		fmt.Sprintf("`%v`", data.ID),
		false,
	)

	embed.AddField(
		"⋆˚꩜｡ Canal de envio",
		fmt.Sprintf("<#%v>", channelID),
		false,
	)

	embed.AddField(
		"⋆˚꩜｡ Publicado por",
		fmt.Sprintf("<@%v>", postAuthor),
		false,
	)

	embed.AddField(
		"⋆˚꩜｡ Status",
		"`Publicada com sucesso!`",
		false,
	)

	return embed.Build()
}

func EmbedUpdateEmbeds(thumbnail string, updateAuthor string, data domain.Template, source string) messages.Embed {
	embed := messages.NewEmbedBuilder().
		SetTitle("⋆. 𐙚˚࿔ Atualizar uma Incorporação 𝜗𝜚˚⋆").
		SetDescription("Sua incorporação foi atualizada com sucesso! ♡").
		SetThumbnail(thumbnail).
		SetColor(0xB4E4D6).
		SetTimestampNow()

	embed.AddField(
		"✏️ Identificador",
		fmt.Sprintf("`%v`", data.ID),
		false,
	)

	embed.AddField(
		"⋆˚꩜｡ Tipo de envio",
		fmt.Sprintf("`%v`", source),
		false,
	)

	embed.AddField(
		"⋆˚꩜｡ Atualizado por",
		fmt.Sprintf("<@%v>", updateAuthor),
		false,
	)

	embed.AddField(
		"⋆˚꩜｡ Status",
		"`Atualizada com sucesso!`",
		false,
	)

	return embed.Build()
}

func EmbedDeleteEmbeds(thumbnail string, removedAuthor string, id int) messages.Embed {
	embed := messages.NewEmbedBuilder().
		SetTitle("⋆. 𐙚˚࿔ Deletar uma Incorporação 𝜗𝜚˚⋆").
		SetDescription("Sua incorporação foi removida com sucesso! ♡").
		SetThumbnail(thumbnail).
		SetColor(0xB4E4D6).
		SetTimestampNow()

	embed.AddField(
		"🗑️ Identificador",
		fmt.Sprintf("`%v`", id),
		false,
	)

	embed.AddField(
		"⋆˚꩜｡ Removido por",
		fmt.Sprintf("<@%v>", removedAuthor),
		false,
	)

	embed.AddField(
		"⋆˚꩜｡ Status",
		"`Removida com sucesso!`",
		false,
	)

	return embed.Build()
}

func EmbedCreateAutoReaction(thumbnail string, data domain.AutoReactions) messages.Embed {
	embed := messages.NewEmbedBuilder().
		SetTitle("⋆. 𐙚˚࿔ Criação de AutoReaction 𝜗𝜚˚⋆").
		SetThumbnail(thumbnail).
		SetColor(0xB4E4D6).
		SetTimestampNow()

	embed.AddField(
		"── .✦ Tipo de monitoramento:",
		fmt.Sprintf("`%v`", data.Type),
		false,
	)

	embed.AddField(
		"⋆˚꩜｡ Canal registrado:",
		fmt.Sprintf("<#%v>", data.ChannelID),
		false,
	)

	embed.AddField(
		"⋆˚꩜｡ Emoji:",
		fmt.Sprintf("%v", data.Emoji),
		false,
	)

	return embed.Build()
}

func EmbedObtainAutoReaction(thumbnail string, data domain.AutoReactions) messages.Embed {
	embed := messages.NewEmbedBuilder().
		SetTitle("⋆. 𐙚˚࿔ Obter uma AutoReaction 𝜗𝜚˚⋆").
		SetThumbnail(thumbnail).
		SetColor(0xB4E4D6).
		SetTimestampNow()

	embed.AddField(
		"── .✦ Tipo de monitoramento:",
		fmt.Sprintf("`%v`", data.Type),
		false,
	)

	embed.AddField(
		"⋆˚꩜｡ Canal registrado:",
		fmt.Sprintf("<#%v>", data.ChannelID),
		false,
	)

	embed.AddField(
		"⋆˚꩜｡ Emoji:",
		fmt.Sprintf("%v", data.Emoji),
		false,
	)

	return embed.Build()
}

func EmbedObtainAllAutoReactions(thumbnail string, data []domain.AutoReactions) messages.Embed {
	embed := messages.NewEmbedBuilder().
		SetTitle("⋆. 𐙚˚࿔ Obter todas AutoReactions 𝜗𝜚˚⋆").
		SetThumbnail(thumbnail).
		SetColor(0xB4E4D6).
		SetTimestampNow()

	for _, value := range data {
		embed.AddField(
			fmt.Sprintf("<#%v> -> %v", value.ChannelID, value.Emoji),
			fmt.Sprintf("Tipo de monitoramento: `%v`", value.Type),
			false,
		)
	}

	return embed.Build()
}

func EmbedUpdateAutoReaction(thumbnail string, data domain.AutoReactions) messages.Embed {
	embed := messages.NewEmbedBuilder().
		SetTitle("⋆. 𐙚˚࿔ Atualizar uma AutoReaction 𝜗𝜚˚⋆").
		SetThumbnail(thumbnail).
		SetColor(0xB4E4D6).
		SetTimestampNow()

	embed.AddField(
		"── .✦ Tipo de monitoramento:",
		fmt.Sprintf("`%v`", data.Type),
		false,
	)

	embed.AddField(
		"⋆˚꩜｡ Canal registrado:",
		fmt.Sprintf("<#%v>", data.ChannelID),
		false,
	)

	embed.AddField(
		"⋆˚꩜｡ Emoji:",
		fmt.Sprintf("%v", data.Emoji),
		false,
	)

	return embed.Build()
}

func EmbedDeleteAutoReaction(thumbnail string, data domain.AutoReactions) messages.Embed {
	embed := messages.NewEmbedBuilder().
		SetTitle("⋆. 𐙚˚࿔ Deletar uma AutoReaction 𝜗𝜚˚⋆").
		SetThumbnail(thumbnail).
		SetColor(0xB4E4D6).
		SetTimestampNow()

	embed.AddField(
		"── .✦ Tipo de monitoramento:",
		fmt.Sprintf("`%v`", data.Type),
		false,
	)

	embed.AddField(
		"⋆˚꩜｡ Canal registrado:",
		fmt.Sprintf("<#%v>", data.ChannelID),
		false,
	)

	embed.AddField(
		"⋆˚꩜｡ Emoji:",
		fmt.Sprintf("%v", data.Emoji),
		false,
	)

	return embed.Build()
}

func EmbedThrowTomato() messages.Embed {
	url := []string{
		"https://static2.klipy.com/ii/4493325008d34b7bf8cd6813cd5c1619/b5/9d/0xfBEUsakgNcBY9qp.gif",
		"https://static2.klipy.com/ii/4493325008d34b7bf8cd6813cd5c1619/a4/95/zCaGxb1H505JWj.gif",
		"https://static2.klipy.com/ii/4493325008d34b7bf8cd6813cd5c1619/9c/04/5qUyahRM8BTCI3DZMh.gif",
	}
	return messages.NewEmbedBuilder().
		SetColor(0xff3333).
		SetTimestampNow().
		SetImage(url[rand.IntN(len(url))]).
		Build()
}
