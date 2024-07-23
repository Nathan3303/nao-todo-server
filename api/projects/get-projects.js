const Project = require("../../models/project");
const serialExecution = require("../../utils/serial-execution");
const buildResponseData = require("../../utils/build-response-data");

module.exports = async (request, response, _p) => {
    const {
        userId,
        title,
        isArchived,
        isDeleted,
        isFinished,
        isFavorite,
        page,
        limit,
        sort,
    } = request.query;

    const tasks = [
        () => _p.handleUserId(userId),
        () => _p.handleTitle(title),
        () => _p.handleIsDeleted(false),
        () => _p.handleIsFinished(isFinished),
        () => _p.handleIsArchived(isArchived),
        () => _p.handlePage(page, limit),
        () => _p.handleSort(sort),
        () => _p.handleOutput(),
    ];

    try {
        let executeResults = await serialExecution(tasks);
        executeResults = executeResults.flat();
        // console.log("executeResults", executeResults);
        const projects = await Project.aggregate(executeResults);
        response.status(200).json(buildResponseData.success(projects));
    } catch (error) {
        response.status(500);
    }
};
