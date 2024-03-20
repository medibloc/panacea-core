# PNFT
PNFT stands for Panacea Non-Fungible Token, an implementation of Cosmos's NFT standards.

## Overview
PNFT allows developers to record and manage unique assets on the blockchain. In this section, we'll introduce the primary functionalities provided by the PNFT system.

### Creating a Denom
A Denom serves as a unique identifier for categorizing NFTs. The MsgCreateDenomRequest structure below is used to create a new Denom. The Creator field represents the account information of the user creating the Denom. The creator of a Denom has the authority to issue NFTs under it.

```go
type MsgCreateDenomRequest struct {
	Id          string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Name        string `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
	Symbol      string `protobuf:"bytes,3,opt,name=symbol,proto3" json:"symbol,omitempty"`
	Description string `protobuf:"bytes,4,opt,name=description,proto3" json:"description,omitempty"`
	Uri         string `protobuf:"bytes,5,opt,name=uri,proto3" json:"uri,omitempty"`
	UriHash     string `protobuf:"bytes,6,opt,name=uri_hash,json=uriHash,proto3" json:"uri_hash,omitempty"`
	Data        string `protobuf:"bytes,7,opt,name=data,proto3" json:"data,omitempty"`
	Creator     string `protobuf:"bytes,8,opt,name=creator,proto3" json:"creator,omitempty"`
}
```


### Updating a Denom
Denom information can be updated using the MsgUpdateDenomRequest structure. All fields except the Id can be modified. The Updater field should include the account information of the user requesting the update.

```go
type MsgUpdateDenomRequest struct {
	Id          string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Name        string `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
	Symbol      string `protobuf:"bytes,3,opt,name=symbol,proto3" json:"symbol,omitempty"`
	Description string `protobuf:"bytes,4,opt,name=description,proto3" json:"description,omitempty"`
	Uri         string `protobuf:"bytes,5,opt,name=uri,proto3" json:"uri,omitempty"`
	UriHash     string `protobuf:"bytes,6,opt,name=uri_hash,json=uriHash,proto3" json:"uri_hash,omitempty"`
	Data        string `protobuf:"bytes,7,opt,name=data,proto3" json:"data,omitempty"`
	Updater     string `protobuf:"bytes,8,opt,name=updater,proto3" json:"updater,omitempty"`
}
```

### Transferring Denom Ownership
To change the ownership of a Denom, use the MsgTransferDenomRequest structure. The Sender should include the account information of the current owner, and the Receiver should include the account information of the new owner. After the transfer, the original owner can no longer issue NFTs under this Denom.

```go
type MsgTransferDenomRequest struct {
	Id       string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Sender   string `protobuf:"bytes,2,opt,name=sender,proto3" json:"sender,omitempty"`
	Receiver string `protobuf:"bytes,3,opt,name=receiver,proto3" json:"receiver,omitempty"`
}
```



### Minting PNFT

This structure allows for the issuance of NFTs within Panacea. The Creator field must contain the account information of the transaction signer. The creator becomes the initial owner and issuer of this NFT. It's important to note that the information of an NFT cannot be altered once created.

```go
type MsgMintPNFTRequest struct {
	DenomId     string `protobuf:"bytes,1,opt,name=denom_id,json=denomId,proto3" json:"denom_id,omitempty"`
	Id          string `protobuf:"bytes,2,opt,name=id,proto3" json:"id,omitempty"`
	Name        string `protobuf:"bytes,3,opt,name=name,proto3" json:"name,omitempty"`
	Description string `protobuf:"bytes,4,opt,name=description,proto3" json:"description,omitempty"`
	Uri         string `protobuf:"bytes,5,opt,name=uri,proto3" json:"uri,omitempty"`
	UriHash     string `protobuf:"bytes,6,opt,name=uri_hash,json=uriHash,proto3" json:"uri_hash,omitempty"`
	Data        string `protobuf:"bytes,7,opt,name=data,proto3" json:"data,omitempty"`
	Creator     string `protobuf:"bytes,8,opt,name=creator,proto3" json:"creator,omitempty"`
}
```


#### Transferring PNFT

This request changes the ownership of an NFT. The Sender must include the account information of the current owner, who also needs to be the one signing the transaction. The Receiver is the new intended owner of the NFT.

```go
type MsgTransferPNFTRequest struct {
	DenomId  string `protobuf:"bytes,1,opt,name=denom_id,json=denomId,proto3" json:"denom_id,omitempty"`
	Id       string `protobuf:"bytes,2,opt,name=id,proto3" json:"id,omitempty"`
	Sender   string `protobuf:"bytes,3,opt,name=sender,proto3" json:"sender,omitempty"`
	Receiver string `protobuf:"bytes,4,opt,name=receiver,proto3" json:"receiver,omitempty"`
}
```


### Burning PNFT

This function allows for the destruction of an NFT. The Burner, who must be the current owner and the one signing the transaction, initiates the burn process. Once an NFT is burned, it cannot be recovered.

```go
type MsgBurnPNFTRequest struct {
	DenomId string `protobuf:"bytes,1,opt,name=denom_id,json=denomId,proto3" json:"denom_id,omitempty"`
	Id      string `protobuf:"bytes,2,opt,name=id,proto3" json:"id,omitempty"`
	Burner  string `protobuf:"bytes,3,opt,name=burner,proto3" json:"burner,omitempty"`
}
```