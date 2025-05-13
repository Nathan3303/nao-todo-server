import { Todo } from '@nao-todo-server/models';
import {
    getJWTPayload,
    useErrorResponseData,
    useResponseData,
    useSuccessfulResponseData
} from '@nao-todo-server/hooks';
import { ObjectId } from '@nao-todo-server/utils';
import type { Request, Response } from 'express';
import type { TodoDueDate } from './types';
import moment from 'moment';

const parseTodoDueDate = (dueDate: TodoDueDate | undefined): TodoDueDate | undefined => {
    const result: TodoDueDate = { startAt: null, endAt: null };
    let _moment;
    if (!dueDate) return void 0;
    if (Object.prototype.hasOwnProperty.call(dueDate, 'startAt')) {
        _moment = moment(dueDate.startAt);
        if (_moment.isValid()) {
            result.startAt = _moment.toDate();
        }
    }
    if (Object.prototype.hasOwnProperty.call(dueDate, 'endAt')) {
        _moment = moment(dueDate.endAt);
        if (_moment.isValid()) {
            result.endAt = _moment.toDate();
        }
    }
    return result;
};

const buildUpdateOptions = (updateInfo: Record<string, unknown>) => {
    const updateOptions = {
        projectId: updateInfo.projectId
            ? new ObjectId(updateInfo.projectId as string)
            : void 0,
        name: updateInfo.name || void 0,
        description: updateInfo.description,
        state: updateInfo.state || void 0,
        priority: updateInfo.priority || void 0,
        dueDate: parseTodoDueDate(updateInfo.dueDate as TodoDueDate),
        isFavorited: typeof updateInfo.isFavorited === 'boolean' ? updateInfo.isFavorited : void 0,
        isArchived: typeof updateInfo.isArchived === 'boolean' ? updateInfo.isArchived : void 0,
        isDeleted: typeof updateInfo.isDeleted === 'boolean' ? updateInfo.isDeleted : void 0,
        isGivenUp: typeof updateInfo.isGivenUp === 'boolean' ? updateInfo.isGivenUp : void 0,
        deletedAt: updateInfo.deletedAt,
        tags: Array.isArray(updateInfo.tags) ? updateInfo.tags.length > 0 ? updateInfo.tags : [] : void 0
    };
    console.log('updateOptions', updateOptions);
    return updateOptions;
};
const updateTodo = async (req: Request, res: Response) => {
    try {
        // 获取请求头中的用户 ID
        const userId = getJWTPayload(req.headers.authorization as string).userId as string;

        // 判断请求参数是否合法
        if (!userId || !req.query.todoId) {
            res.json(useErrorResponseData('参数错误，请求无效'));
            return;
        }

        // 构建更新选项
        const updateOptions = buildUpdateOptions(req.body);

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
        const updateOptions = buildUpdateOptions(req.body.updateInfo);

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
