const pipelines = require("../../pipelines/project");
const buildRD = require("../../utils/build-response-data");
const getProjects = require("./get-projects");

module.exports = async (request, response) => {
    const { method } = request;

    switch (method) {
        case "GET":
            await getProjects(request, response, pipelines);
            break;
        default:
            response.status(200).json(buildRD.error("Invalid method."));
            break;
    }
};
