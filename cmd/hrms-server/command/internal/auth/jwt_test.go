package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"
)

const (
	testRSAPrivateKey = `-----BEGIN PRIVATE KEY-----
MIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQDQPnxDqyoqarif
RrPzxWBSiLEOGZNngv63SkSDicmEod9CaEf6v/9GtghI5NjFNvQW5AFB+SPk3eyF
vRecPPsQLWyxwjdXublrIizulsidzVQt3+8TtdWBTg17vuMrlhckY/XGT6WCpT+w
UfLOK0b6eIKbopF5ReBuXQ6VSQzd4lnOeGNhgnp3PpYb3gYw9bK7P2NXvMSaQukQ
1Sm5B4hAXbKMTI8ViezmutXKxAyBcFBHYZ868QkOY2SFhrOj5AkBASu06KEEkxpS
HG+6v8MaZYozc2lNQnjiEmf03Rn3cVfB4rxNpWAuh4Gq9dibsfgRr8j7cfs27Obg
IuGpKd4FAgMBAAECggEADi0IGIiBCPzVhJeI4yyp2fR5g77fTJ9GPQP5QLBoiiGF
fpJ1aURuBEpFVhuosK7aVD9BNqnXH7AGdzQ1dJVSIfp1o8Qs2vPxrgatjbBbXSKq
ev+mLsP6EiCrb6PBoyicerBd/W9T57OIGmCkaWnYyG7btUHj1UrvHumGzJ15xWuR
Gm9UZRSmE2VN5bKX65xWFmHiGcWccMlnZuyPG9OfhEn4MzXvQwMZ4J9iC8RvmZJg
7qJdEE2iqoSccb3x83Xp+uCphbUHAFKPgalrAN0FIACGJkzTZ4YoYDeTtTiQ5Tq7
VTXULJBOiASfi83+U+LwbZuhAi1GjccsGKHtIwq9AQKBgQDsZAEP10I5mEG02vQ3
UfLlD1iChBudDrMGQYUdeU7E/SxdPoJGy0596jgrcA100EkXK8zvCwrgnzgo7ska
tMy2CS0jgjSZuQQZOEq/sVm0k5xcZXT8bHWULNmYmyhmDM/1AYx1c8tQWsiYu6f9
8xFwpKOmKJ/f8tcnAxeNsaaqyQKBgQDhhMKtuI5Dxp3vKLVOcEmNtO6AXkfk68gI
bv1gsWknfqGygD2EOApXDgkD/tYS/HpSTDRhYe8rLgZDBzmi89eph/kdttdpYQtr
Da116WqcFiVm8rNZMfasVlpm3pY1CaNXvMA9Of8RVrOcjdHJaanr7YDQx0ovItt+
sFLDO3W7XQKBgQCqk+M8Rg2Qt/C6C8FsZeMLLU6mJ6Qxahj/K6pdwVp4xWQNCP1D
DpPeQnQzzBC5uU70vHOODv7TZbFFwEE31z1dIjQDSoKgZqSxejBeMSDVMCsFdWS8
fZs+yDpgZ534PciWOH7dhigxHMFhjRBFLO/pw7QfQ3NSS867ZPzLD2WAGQKBgDm8
7sbhaHMLx+WyS3EQqJRCTYnKGagPgcA/AloeMejtr+JumNFgM62EJ2TBeveTcpHd
ds+z7jLk7q98ixIgUgfSi0JDTLVrJiw7bTyyDRx3Qw4vdyGP/DK1TSHnPRfJJuvQ
pHtIfPhodUXQvXROvDVuMjvBukmFKCMwa5AWihb1AoGAStOWRKodmm+M9Up8xT1Q
BztV+TPEaFcT276H8Ie3EgSNACo0DO9egur1yR3JFJixOTZKCynSOKMHE0PP9TsS
9PW5oSiZKf44G1ksGWhCaOXHIJ22avMJ6LE2fgRcNEUMkoHcy3k//JxoumghDk9b
8XTwmb0kBDFCkU+nQmuENYU=
-----END PRIVATE KEY-----`

	testRSAPublicKey = `-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA0D58Q6sqKmq4n0az88Vg
UoixDhmTZ4L+t0pEg4nJhKHfQmhH+r//RrYISOTYxTb0FuQBQfkj5N3shb0XnDz7
EC1sscI3V7m5ayIs7pbInc1ULd/vE7XVgU4Ne77jK5YXJGP1xk+lgqU/sFHyzitG
+niCm6KReUXgbl0OlUkM3eJZznhjYYJ6dz6WG94GMPWyuz9jV7zEmkLpENUpuQeI
QF2yjEyPFYns5rrVysQMgXBQR2GfOvEJDmNkhYazo+QJAQErtOihBJMaUhxvur/D
GmWKM3NpTUJ44hJn9N0Z93FXweK8TaVgLoeBqvXYm7H4Ea/I+3H7Nuzm4CLhqSne
BQIDAQAB
-----END PUBLIC KEY-----`
)

func TestGenerateTestTokenRejectsHS256(t *testing.T) {
	t.Parallel()

	mw := NewJWTMiddlewareWithOptions("cube-secret", "issuer", "audience", Options{Alg: "HS256"})
	_, err := mw.GenerateTestToken("dev-user", "tenant-123", []string{"ADMIN"}, time.Hour)
	require.Error(t, err)
}

func TestGenerateTestTokenRS256(t *testing.T) {
	t.Parallel()

	mw := NewJWTMiddlewareWithOptions("ignored", "issuer", "audience", Options{
		Alg:           "RS256",
		PublicKeyPEM:  []byte(testRSAPublicKey),
		PrivateKeyPEM: []byte(testRSAPrivateKey),
		KeyID:         "bff-key-1",
	})

	token, err := mw.GenerateTestToken("user-rs", "tenant-rs", []string{"ADMIN", "USER"}, 2*time.Hour)
	require.NoError(t, err)

	claims, err := mw.ValidateToken(token)
	require.NoError(t, err)
	require.Equal(t, "user-rs", claims.UserID)
	require.Equal(t, "tenant-rs", claims.TenantID)
	require.ElementsMatch(t, []string{"ADMIN", "USER"}, claims.Roles)

	parsed, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return mw.publicKey, nil
	})
	require.NoError(t, err)
	require.Equal(t, "bff-key-1", parsed.Header["kid"])
}

func TestGenerateTestTokenRS256MissingPrivateKey(t *testing.T) {
	t.Parallel()

	mw := NewJWTMiddlewareWithOptions("ignored", "issuer", "audience", Options{
		Alg:          "RS256",
		PublicKeyPEM: []byte(testRSAPublicKey),
		KeyID:        "bff-key-1",
	})

	_, err := mw.GenerateTestToken("user", "tenant", []string{"ADMIN"}, time.Minute)
	require.Error(t, err)
}
