const Tag = require("../../models/tag");
const serialExecution = require("../../utils/serial-execution");
const buildResponseData = require("../../utils/build-response-data");

module.exports = async (request, response, _p) => {
    const { userId, name, isDeleted, page, limit } = request.query;

    const tasks = [
        () => _p.handleUserId(userId),
        () => _p.handleName(name),
        () => _p.handleIsDeleted(isDeleted),
        () => _p.handlePage(page, limit),
        () => _p.handleSelectFields(),
        () => Tag.aggregate().pipeline(),
    ];

    try {
        let executeResults = await serialExecution(tasks);
        executeResults = executeResults.flat();
        const tags = await Tag.aggregate(executeResults);
        response.status(200).json(buildResponseData.success(tags));
    } catch (error) {
        response.status(200).json(buildResponseData.error(error.message));
    }
};
