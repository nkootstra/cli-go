FROM node:22.12-alpine

RUN npm install --global oauth2-mock-server

EXPOSE 8080

CMD ["oauth2-mock-server", "-a", "0.0.0.0", "-p", "8080"]
