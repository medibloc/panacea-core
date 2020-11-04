package types

import "encoding/json"

type Issuance = MsgIssueCurrency

func (i Issuance) Empty() bool {
	return i.IssuerAddress.Empty()
}

func (i Issuance) Valid() bool {
	return i.ValidateBasic() == nil
}

func (i Issuance) String() string {
	bz, _ := json.Marshal(i)
	return string(bz)
}
