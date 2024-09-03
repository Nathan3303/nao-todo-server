const pipelines = require("../../pipelines/tag");
const buildRD = require("../../utils/build-response-data");
const getTags = require("./get-tags");

module.exports = async (request, response) => {
    const { method } = request;

    switch (method) {
        case "GET":
            await getTags(request, response, pipelines);
            break;
        default:
            response.status(200).json(buildRD.error("Invalid method."));
            break;
    }
};
