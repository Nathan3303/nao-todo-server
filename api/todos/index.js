const buildRD = require("../../utils/build-response-data");

const getTodos = require("./get-todos");
const pipelines = require("./pipelines");

module.exports = async (request, response) => {
    // Distribute request to corresponding method
    switch (request.method) {
        case "GET":
            await getTodos(request, response, pipelines);
            break;
        default:
            response.status(200).json(buildRD.error("Invalid method."));
            break;
    }
};
