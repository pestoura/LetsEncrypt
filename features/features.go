//go:generate stringer -type=FeatureFlag

package features

import (
	"fmt"
	"strings"
	"sync"
)

type FeatureFlag int

const (
	unused FeatureFlag = iota // unused is used for testing

	//   Currently in-use features
	// Check CAA and respect validationmethods parameter.
	CAAValidationMethods
	// Check CAA and respect accounturi parameter.
	CAAAccountURI
	// EnforceMultiVA causes the VA to block on remote VA PerformValidation
	// requests in order to make a valid/invalid decision with the results.
	EnforceMultiVA
	// MultiVAFullResults will cause the main VA to wait for all of the remote VA
	// results, not just the threshold required to make a decision.
	MultiVAFullResults
	// MandatoryPOSTAsGET forbids legacy unauthenticated GET requests for ACME
	// resources.
	MandatoryPOSTAsGET
	// ECDSAForAll enables all accounts, regardless of their presence in the CA's
	// ecdsaAllowedAccounts config value, to get issuance from ECDSA issuers.
	ECDSAForAll
	// ServeRenewalInfo exposes the renewalInfo endpoint in the directory and for
	// GET requests. WARNING: This feature is a draft and highly unstable.
	ServeRenewalInfo
	// AllowReRevocation causes the RA to allow the revocation reason of an
	// already-revoked certificate to be updated to `keyCompromise` from any
	// other reason if that compromise is demonstrated by making the second
	// revocation request signed by the certificate keypair.
	AllowReRevocation
	// MozRevocationReasons causes the RA to enforce the following upcoming
	// Mozilla policies regarding revocation:
	// - A subscriber can request that their certificate be revoked with reason
	//   keyCompromise, even without demonstrating that compromise at the time.
	//   However, the cert's pubkey will not be added to the blocked keys list.
	// - When an applicant other than the original subscriber requests that a
	//   certificate be revoked (by demonstrating control over all names in it),
	//   the cert will be revoked with reason cessationOfOperation, regardless of
	//   what revocation reason they request.
	// - When anyone requests that a certificate be revoked by signing the request
	//   with the certificate's keypair, the cert will be revoked with reason
	//   keyCompromise, regardless of what revocation reason they request.
	MozRevocationReasons
	// OldTLSOutbound allows the VA to negotiate TLS 1.0 and TLS 1.1 during
	// HTTPS redirects. When it is set to false, the VA will only connect to
	// HTTPS servers that support TLS 1.2 or above.
	OldTLSOutbound
	// OldTLSInbound controls whether the WFE rejects inbound requests using
	// TLS 1.0 and TLS 1.1. Because WFE does not terminate TLS in production,
	// we rely on the TLS-Version header (set by our reverse proxy).
	OldTLSInbound
	// SHA1CSRs controls whether the /acme/finalize endpoint rejects CSRs that
	// are self-signed using SHA1.
	SHA1CSRs
	// AllowUnrecognizedFeatures is internal to the features package: if true,
	// skip error when unrecognized feature flag names are passed.
	AllowUnrecognizedFeatures
)

// List of features and their default value, protected by fMu
var features = map[FeatureFlag]bool{
	unused:                    false,
	CAAValidationMethods:      false,
	CAAAccountURI:             false,
	EnforceMultiVA:            false,
	MultiVAFullResults:        false,
	MandatoryPOSTAsGET:        false,
	ECDSAForAll:               false,
	ServeRenewalInfo:          false,
	AllowReRevocation:         false,
	MozRevocationReasons:      false,
	OldTLSOutbound:            true,
	OldTLSInbound:             true,
	SHA1CSRs:                  true,
	AllowUnrecognizedFeatures: false,
}

var fMu = new(sync.RWMutex)

var initial = map[FeatureFlag]bool{}

var nameToFeature = make(map[string]FeatureFlag, len(features))

func init() {
	for f, v := range features {
		nameToFeature[f.String()] = f
		initial[f] = v
	}
}

// Set accepts a list of features and whether they should
// be enabled or disabled. In the presence of unrecognized
// flags, it will return an error or not depending on the
// value of AllowUnrecognizedFeatures.
func Set(featureSet map[string]bool) error {
	fMu.Lock()
	defer fMu.Unlock()
	var unknown []string
	for n, v := range featureSet {
		f, present := nameToFeature[n]
		if present {
			features[f] = v
		} else {
			unknown = append(unknown, n)
		}
	}
	if len(unknown) > 0 && !features[AllowUnrecognizedFeatures] {
		return fmt.Errorf("unrecognized feature flag names: %s",
			strings.Join(unknown, ", "))
	}
	return nil
}

// Enabled returns true if the feature is enabled or false
// if it isn't, it will panic if passed a feature that it
// doesn't know.
func Enabled(n FeatureFlag) bool {
	fMu.RLock()
	defer fMu.RUnlock()
	v, present := features[n]
	if !present {
		panic(fmt.Sprintf("feature '%s' doesn't exist", n.String()))
	}
	return v
}

// Reset resets the features to their initial state
func Reset() {
	fMu.Lock()
	defer fMu.Unlock()
	for k, v := range initial {
		features[k] = v
	}
}
