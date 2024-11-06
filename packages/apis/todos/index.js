const pipelines = require("../../pipelines/todo");
const buildRD = require("../../utils/build-response-data");
const getTodos = require("./get-todos");
const updateTodos = require("./update-todos");

module.exports = async (request, response) => {
    switch (request.method) {
        case "GET":
            await getTodos(request, response, pipelines);
            break;
        case "PUT":
            await updateTodos(request, response, pipelines);
            break;
        default:
            response.status(200).json(buildRD.error("Invalid method."));
            break;
    }
};