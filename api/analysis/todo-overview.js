const Todo = require("../../models/todo");
const buildRD = require("../../utils/build-response-data");
const moment = require("moment");
const ObjectId = require("mongoose").Types.ObjectId;

module.exports = async (request, response) => {
    const { projectId } = request.query;

    if (!projectId) {
        response.status(200).json(buildRD.error("projectId is required"));
        return;
    }

    try {
        const now = moment().utcOffset(8);
        const today = now.toDate();
        const pass15days = now.subtract(15, "days").toDate();
        // console.log(projectId);
        const pipeline = [
            { $match: { projectId: new ObjectId(projectId) } },
            { $match: { createdAt: { $gt: pass15days, $lte: today } } },
            {
                $group: {
                    _id: {
                        date: {
                            $dateToString: {
                                format: "%Y-%m-%d",
                                date: "$createdAt",
                            },
                        },
                        state: "$state",
                    },
                    count: { $sum: 1 },
                },
            },
            {
                $group: {
                    _id: "$_id.date",
                    states: {
                        $push: {
                            state: "$_id.state",
                            count: "$count",
                        },
                    },
                },
            },
            { $sort: { _id: 1 } },
        ];
        // console.log(pipeline);
        const todosInMonth = await Todo.aggregate(pipeline).exec();
        response.status(200).json(buildRD.success(todosInMonth));
    } catch (error) {
        response.status(200).json(buildRD.error(error.message));
    }
};
