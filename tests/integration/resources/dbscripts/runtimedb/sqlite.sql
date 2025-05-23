-- Table to store OAuth2 authorization codes.
CREATE TABLE IDN_OAUTH2_AUTHZ_CODE (
    ID INTEGER PRIMARY KEY AUTOINCREMENT,
    CODE_ID VARCHAR(36) UNIQUE NOT NULL,
    AUTHORIZATION_CODE VARCHAR(500) NOT NULL,
    CONSUMER_KEY VARCHAR(255) NOT NULL,
    CALLBACK_URL VARCHAR(500),
    AUTHZ_USER VARCHAR(255) NOT NULL,
    TIME_CREATED DATETIME NOT NULL,
    EXPIRY_TIME DATETIME NOT NULL,
    STATE VARCHAR(50) NOT NULL
);

-- Table to store scopes associated with OAuth2 authorization codes.
CREATE TABLE IDN_OAUTH2_AUTHZ_CODE_SCOPE (
    ID INTEGER PRIMARY KEY AUTOINCREMENT,
    CODE_ID VARCHAR(36) NOT NULL REFERENCES IDN_OAUTH2_AUTHZ_CODE(CODE_ID) ON DELETE CASCADE,
    SCOPE VARCHAR(255) NOT NULL,
    UNIQUE (CODE_ID, SCOPE)
);
