const buildRD = require("../../utils/build-response-data");
const pipelines = require("../../pipelines/event");
const getEvents = require("./get-events");

module.exports = async function (request, response) {
    switch (request.method) {
        case "GET":
            await getEvents(request, response, pipelines);
            break;
        default:
            response.status(200).json(buildRD.error("Method not allowed"));
            break;
    }
};
