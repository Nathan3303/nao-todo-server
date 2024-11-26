import { Todo } from '@nao-todo-server/models';
import {
    getJWTPayload,
    useErrorResponseData,
    useResponseData,
    useSuccessfulResponseData
} from '@nao-todo-server/hooks';
import { ObjectId } from '@nao-todo-server/utils';
import type { Request, Response } from 'express';
import moment from 'moment/moment';

const buildUpdateOptions = (req: Request) => {
    // 定义更新选项
    const startAt = req.body.dueDate.startAt as string;
    const endAt = req.body.dueDate.endAt as string;
    const updateOptions = {
        projectId: req.body.projectId
            ? new ObjectId(req.body.projectId as string)
            : void 0,
        name: req.body.name || void 0,
        description: req.body.description || void 0,
        state: req.body.state || void 0,
        priority: req.body.priority || void 0,
        dueDate: {
            startAt: startAt ? moment(startAt).utc(true) : void 0,
            endAt: endAt ? moment(endAt).utc(true) : void 0
        },
        isFavorited: req.body.isFavorited || void 0,
        isArchived: req.body.isArchived || void 0,
        isDeleted: req.body.isDeleted || void 0,
        tags: req.body.tags || void 0
    };
    // 移除未定义的值
    Object.keys(updateOptions).forEach(key => {
        if (updateOptions[key as keyof typeof updateOptions] === void 0) {
            delete updateOptions[key as keyof typeof updateOptions];
        }
    });
    // 返回
    return updateOptions;
};

const updateTodo = async (req: Request, res: Response) => {
    try {
        // 获取请求头中的用户 ID
        const userId = getJWTPayload(req.headers.authorization as string)
            .userId as string;

        // 判断请求参数是否合法
        if (!userId || !req.query.todoId) {
            res.json(useErrorResponseData('参数错误，请求无效'));
            return;
        }

        // 构建更新选项
        const updateOptions = buildUpdateOptions(req);

        // 更新数据
        const updatedTodo = await Todo.findOneAndUpdate(
            {
                _id: new ObjectId(req.query.todoId as string),
                userId: new ObjectId(userId)
            },
            { $set: updateOptions },
            { new: true }
        ).exec();

        // 判断更新结果，返回对应的 JSON 数据
        if (updatedTodo) {
            res.json(
                useSuccessfulResponseData({
                    todoId: updatedTodo._id.toString()
                })
            );
        } else {
            res.json(useErrorResponseData('更新失败'));
            return;
        }
    } catch (error) {
        console.log('[api/todo/updateTodo]', error);
        res.json(useResponseData(50001, '更新失败，请稍后再试', null));
    }
};

const updateTodos = async (req: Request, res: Response) => {
    try {
        // 获取请求头中的用户 ID
        const userId = getJWTPayload(req.headers.authorization as string)
            .userId as string;

        // 判断请求参数是否合法
        if (!userId || !req.body.todoIds || !req.body.updateInfo) {
            res.json(useErrorResponseData('参数错误，请求无效'));
            return;
        }

        // 构建更新选项
        const updateTodoIds = req.body.todoIds.map(
            (todoId: string) => new ObjectId(todoId)
        );
        const updateOptions = buildUpdateOptions(req);

        // 更新数据
        const updateResult = await Todo.updateMany(
            { _id: { $in: updateTodoIds } },
            { $set: updateOptions },
            { multi: true }
        ).exec();

        // 判断更新结果，返回对应的 JSON 数据
        if (
            updateResult &&
            updateResult.modifiedCount == updateResult.matchedCount
        ) {
            res.json(
                useSuccessfulResponseData({
                    todoIds: req.body.todoIds,
                    total: req.body.todoIds.length,
                    modifiedCount: updateResult.modifiedCount
                })
            );
        } else {
            res.json(useErrorResponseData('更新失败'));
        }
    } catch (error) {
        console.log('[api/todo/updateTodos] Error:', error);
        res.json(useResponseData(50001, '更新失败，请稍候重试', null));
    }
};

export { updateTodo, updateTodos };
