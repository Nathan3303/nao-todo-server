const buildRD = require("../../utils/build-response-data");
const { checkMethod } = require("../../utils");
const getTodoOverview = require("./todo-overview");

module.exports = async function (request, response) {
    if (checkMethod(request, response, "GET")) {
        return;
    }

    const { target } = request.query;

    if (!target) {
        response.status(200).json(buildRD.error("Invalid target"));
    }

    switch (target) {
        case "todo-overview":
            await getTodoOverview(request, response);
            // response.status(200).json(buildRD.success(result));
            return;
        default:
            response.status(200).json(buildRD.error("Invalid target"));
            break;
    }
};
