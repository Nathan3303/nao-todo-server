import { Todo } from '@nao-todo-server/models';
import {
    getJWTPayload,
    useErrorResponseData,
    useResponseData,
    useSuccessfulResponseData
} from '@nao-todo-server/hooks';
import { ObjectId } from '@nao-todo-server/utils';
import type { Request, Response } from 'express';
import moment from 'moment';

const createTodo = async (req: Request, res: Response) => {
    try {
        // 获取请求头中的用户 ID
        const userId = getJWTPayload(req.headers.authorization as string)
            .userId as string;

        // 判断请求参数是否正确
        if (!userId || !req.body.projectId || !req.body.name) {
            res.json(useErrorResponseData('参数错误，请求无效'));
            return;
        }

        // 构建新增选项
        const startAt = req.body.dueDate.startAt as string;
        const endAt = req.body.dueDate.endAt as string;
        const createOptions = {
            userId: new ObjectId(userId),
            projectId: req.body.projectId || void 0,
            name: req.body.name || void 0,
            description: req.body.description || void 0,
            state: req.body.state || void 0,
            priority: req.body.priority || void 0,
            dueDate: {
                startAt: startAt ? moment(startAt).utc(true) : void 0,
                endAt: endAt ? moment(endAt).utc(true) : void 0
            },
            tags: req.body.tags || []
        };

        // 创建数据
        const newTodo = await Todo.create(createOptions);

        // 判断创建结果，并返回对应的 JSON 数据
        if (newTodo) {
            res.json(
                useSuccessfulResponseData({
                    ...newTodo.toJSON(),
                    id: newTodo._id.toString()
                })
            );
        } else {
            res.json(useErrorResponseData('创建失败'));
        }
    } catch (error) {
        console.log('[api/todo/createTodo] Error:', error);
        res.json(useResponseData(50001, '创建失败，请稍后再试', null));
    }
};

const createTodos = async () => {};

export { createTodo, createTodos };
