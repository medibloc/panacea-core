# Token

The Token module, which can issue new tokens.

### Overview

#### Issue a new token

```go
type MsgIssueToken struct {
	Name         Name           `json:"name"`
	Symbol       Symbol         `json:"symbol"`
	TotalSupply  sdk.Int        `json:"total_supply"`
	Mintable     bool           `json:"mintable"`
	OwnerAddress sdk.AccAddress `json:"owner_address"`
}
```

A new token can be issued by the transaction message above. Anyone can issue a new token with fee paid.

Each field in the message has the following limits:

##### `Name`
A human-friendly name, limited to 32 characters, such as `Hello Token`.

##### `Symbol`
An identifier of the token, limited to alphanumeric characters and is case-insensitive, such as `KAI`.
The first letter must be an alphabet character.
The length of symbol should be between 3 and 13.

The symbol doesn't have to be unique. `-` followed by random 3 letters will be appended to the provided symbol to avoid uniqueness constraint.
For example, `KAI-0EA`.
Those 3 letters are the first three letters of the Tx hash of the `issue` transaction.
So, the total length of the generated symbol will be between 7 and 17.
The generated symbol will be returned as a Tx response.

##### `TotalSupply`
An amount of the total supply in micro unit (by 1e6 for decimal part). The max total supply is 90 billion. 

##### `Mintable`
That means whether this token can be minted in the future.

##### `OwnerAddress`
An issuer address of the transaction. It will become the owner of the token. All supplied tokens will be in this account.
