const express = require("express");
const mongoose = require("mongoose");
const cors = require("cors");
const { verifyJWT } = require("./utils/make-jwt");

const app = express();

app.use(cors());
app.use(express.json());
app.use(express.urlencoded({ extended: true }));
app.use(function (request, response, next) {
    const { authorization } = request.headers;
    if (!authorization) {
        response
            .status(200)
            .json(buildRD.error("Authorization header is missing."));
        return;
    }
    const token = authorization.replace("Bearer ", "");
    const verifyResult = verifyJWT(token);
    if (!verifyResult) {
        response.status(200).json(buildRD.error("Invalid token."));
        return;
    }
    // console.log("Authorization success.");
    next();
});

app.get("/api/projects", require("./api/projects"));
app.get("/api/todos", require("./api/todos"));
app.get("/api/analysis", require("./api/analysis"));
app.use("/api/tags", require("./api/tags"));

app.use("/api/project", require("./api/project"));
app.use("/api/todo", require("./api/todo"));
app.use("/api/event", require("./api/event"));
app.use("/api/tag", require("./api/tag"));

app.use("/", (_, res) => res.end("Hello World!"));

mongoose
    .set("strictQuery", true)
    .connect(`mongodb://localhost/naotodo`)
    .then(() => {
        app.listen(3002, () => {
            console.log("NaoTodoServer(dev) is running on port 3002");
        });
    });
