import winston from "winston";

const logger = winston.createLogger({
  format: winston.format.json(),
  level: "info",
});

if (process.env.NODE_ENV !== "production") {
  logger.add(new winston.transports.Console({
    format: winston.format.simple(),
  }));
}

export default logger;
