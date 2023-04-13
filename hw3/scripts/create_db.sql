
DROP TABLE IF EXISTS Users;
CREATE TABLE Users(
    Username String UNIQUE NOT NULL ,
    Email String UNIQUE NOT NULL ,
    -- IsValid INT,
    Password String NOT NULL ,
    Token String
);

DROP TABLE IF EXISTS Verification;
CREATE TABLE Verification(
  Email String UNIQUE NOT NULL ,
  Username String UNIQUE ,
  Password String ,
  Token String
  -- FOREIGN KEY (Email) REFERENCES Users.Email
);