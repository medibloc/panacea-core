---
sidebar_position: 1
---

# AOL

The AOL\(Append Only Log\), which can store various useful data such as healthcare data, is the key feature of the Panacea blockchain. It allows user to manage topics, authorities, and data.

### Overview

#### Create Topic

```go
// MsgCreateTopic - create topics
type MsgCreateTopic struct {
	TopicName    string         `json:"topic_name"`
	Description  string         `json:"description"`
	OwnerAddress sdk.AccAddress `json:"owner_address"`
}
```

A `MsgCreateTopic` is constructed to facilitate the AOL topic. Message sender can set a topic and description. The writer can add record after receiving the appropriate privileges.

Each field has limit. See the [Limits](#limits) section.

#### Add Writer

```go
// MsgAddWriter - add writer to the topic
type MsgAddWriter struct {
	TopicName     string         `json:"topic_name"`
	Moniker       string         `json:"moniker"`
	Description   string         `json:"description"`
	WriterAddress sdk.AccAddress `json:"writer_address"`
	OwnerAddress  sdk.AccAddress `json:"owner_address"`
}
```

The owner of the exist topics can manage writer privileges. Writer who received privileges from the topic owner can add record to the topic. This means that the owner authenticate the writer. This `MsgAddWriter` performs a function similar to issuing a certificate.

Each field has limit. See the [Limits](#limits) section.

#### Delete Writer

```go
// MsgDeleteWriter - delete writer from the topic
type MsgDeleteWriter struct {
	TopicName     string         `json:"topic_name"`
	WriterAddress sdk.AccAddress `json:"writer_address"`
	OwnerAddress  sdk.AccAddress `json:"owner_address"`
}
```

This `MsgDeleteWriter` removes writer from the topic. It is impossible to add record to the topic after being deprived of authority.

Each field has limit. See the [Limits](#limits) section.

#### Add Record

```go
// MsgAddRecord - add record to topic
type MsgAddRecord struct {
	TopicName       string         `json:"topic_name"`
	Key             []byte         `json:"key"`
	Value           []byte         `json:"value"`
	WriterAddress   sdk.AccAddress `json:"writer_address"`
	OwnerAddress    sdk.AccAddress `json:"owner_address"`
	FeePayerAddress sdk.AccAddress `json:"fee_payer_address"`
}
```

This `MsgAddRecord` add record or any data to the topic. If `FeePayerAddress` is provided, node charges fee to FeePayer.

Each field has limit. See the [Limits](#limits) section.

### Limits

|Field|Min Length|Max Length|Charset|
|-----|----------|----------|-------|
|`TopicName`|1|70|`a-z`, `A-Z`, `0-9`, `.`, `_` and `-`|
|`Moniker`|0|70|`a-z`, `A-Z`, `0-9`, `.`, `_` and `-`|
|`Description`|0|5,000|Any|
|`Key`|0|70|Any|
|`Value`|0|5,000|Any|
