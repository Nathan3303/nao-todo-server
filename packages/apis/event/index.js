const buildRD = require("../../utils/build-response-data");
const createEvent = require("./create-event");
const updateEvent = require("./update-event");
const deleteEvent = require("./delete-event");

module.exports = async function (request, response) {
    switch (request.method) {
        case "POST":
            await createEvent(request, response);
            break;
        case "DELETE":
            await deleteEvent(request, response);
            break;
        case "PUT":
            await updateEvent(request, response);
            break;
        default:
            response.status(200).json(buildRD.error("Method not allowed"));
            break;
    }
};
