package agent

import (
	"encoding/json"
	"fmt"

	"github.com/Soliard/go-tpl-metrics/internal/compressor"
	"github.com/Soliard/go-tpl-metrics/internal/crypto"
	"github.com/Soliard/go-tpl-metrics/internal/signer"
)

// prepareJSONPayload подготавливает тело запроса:
// json.Marshal -> EncryptHybrid (если ключ есть) -> gzip.
// Возвращает сжатый буфер и подпись (по сжатому буферу) в base64, если есть signKey.
func (a *Agent) prepareJSONPayload(v any) (compressed []byte, signatureB64 string, err error) {
	buf, err := json.Marshal(v)
	if err != nil {
		return nil, "", fmt.Errorf("cant marshal payload: %v", err)
	}
	if a.hasCryptoKey() {
		enc, err := crypto.EncryptHybrid(buf, a.publicKey)
		if err != nil {
			return nil, "", fmt.Errorf("cant encrypt data: %v", err)
		}
		buf = enc
		a.Logger.Info("payload encrypted successfully")
	}
	comp, err := compressor.CompressData(buf)
	if err != nil {
		return nil, "", fmt.Errorf("cant compress data: %v", err)
	}
	var signB64 string
	if a.hasSignKey() {
		sig := signer.Sign(comp, a.signKey)
		signB64 = signer.EncodeSign(sig)
	}
	return comp, signB64, nil
}
