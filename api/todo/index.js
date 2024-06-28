const getTodo = require("./get-todo");
const createTodo = require("./create-todo");
const deleteTodo = require("./delete-todo");
const updateTodo = require("./update-todo");
const buildRD = require("../../utils/build-response-data");

module.exports = async function (request, response) {
    switch (request.method) {
        case "GET":
            await getTodo(request, response);
            break;
        case "POST":
            await createTodo(request, response);
            break;
        case "DELETE":
            await deleteTodo(request, response);
            break;
        case "PUT":
            await updateTodo(request, response);
            break;
        default:
            response.status(200).json(buildRD.error("Method not allowed"));
            break;
    }
};
