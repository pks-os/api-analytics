import Fastify from "fastify";
import { fastifyAnalytics } from "node-api-analytics";
import * as dotenv from "dotenv";
dotenv.config();

const apiKey = process.env.API_KEY;

const fastify = Fastify({
  logger: true,
});

fastify.addHook("onRequest", fastifyAnalytics(apiKey));

fastify.get("/", function (request, reply) {
  reply.send({ message: "Hello World!" });
});

fastify.listen({ port: 8080 }, function (err, address) {
  console.log("Server listening at http://localhost:8080");
  if (err) {
    fastify.log.error(err);
    process.exit(1);
  }
});
