CREATE TABLE If NOT EXISTS transactions (
    ID UUID PRIMARY KEY,
    TransType VARCHAR(50) NOT NULL,
    Category VARCHAR(100) NOT NULL,
    Amount DECIMAL(15, 2) NOT NULL,
    TransDate TIMESTAMP NOT NULL,
    Description TEXT
)