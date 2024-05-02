USE `retro-games`;

DROP TABLE IF EXISTS `users`;
DROP TABLE IF EXISTS `games`;
DROP TABLE IF EXISTS `offers`;

CREATE TABLE `users` (
  `userId` int NOT NULL AUTO_INCREMENT,
  `email` varchar(255) DEFAULT NULL,
  `name` varchar(255) DEFAULT NULL,
  `address` varchar(255) DEFAULT NULL,
  `password` varchar(255) DEFAULT NULL,
  PRIMARY KEY (`userId`),
  UNIQUE KEY `email` (`email`)
);

CREATE TABLE `games` (
  `gameId` int NOT NULL AUTO_INCREMENT,
  `userId` int NOT NULL,
  `name` varchar(255) DEFAULT NULL,
  `publisher` varchar(255) DEFAULT NULL,
  `year` int DEFAULT NULL,
  `system` varchar(255) DEFAULT NULL,
  `condition` enum('mint','good','fair','poor') DEFAULT NULL,
  `owners` int DEFAULT NULL,
  PRIMARY KEY (`gameId`),
  KEY `userId` (`userId`),
  FOREIGN KEY (`userId`) REFERENCES `users` (`userId`)
);

CREATE TABLE `offers` (
  `offerId` int NOT NULL AUTO_INCREMENT,
  `offererUserId` int NOT NULL,
  `recipientUserId` int NOT NULL,
  `offererGameId` int NOT NULL,
  `recipientGameId` int NOT NULL,
  `status` enum('pending', 'cancelled', 'rejected', 'accepted') DEFAULT 'pending',
  PRIMARY KEY (`offerId`),
  FOREIGN KEY (`offererUserId`) REFERENCES `users` (`userId`),
  FOREIGN KEY (`recipientUserId`) REFERENCES `users` (`userId`),
  FOREIGN KEY (`offererGameId`) REFERENCES `games` (`gameId`),
  FOREIGN KEY (`recipientGameId`) REFERENCES `games` (`gameId`)
);