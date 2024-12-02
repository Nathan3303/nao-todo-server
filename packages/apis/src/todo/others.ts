import { Event, Todo } from '@nao-todo-server/models';
import {
    getJWTPayload,
    useErrorResponseData,
    useResponseData,
    useSuccessfulResponseData
} from '@nao-todo-server/hooks';
import { ObjectId } from '@nao-todo-server/utils';
import type { Request, Response } from 'express';

const duplicateTodo = async (req: Request, res: Response) => {
    try {
        // 获取请求头中的 UserId
        const userId = getJWTPayload(req.headers.authorization as string)
            .userId as string;

        // 判断请求参数是否正确
        if (!userId || !req.query.todoId) {
            res.json(useErrorResponseData('参数错误，请求无效'));
            return;
        }

        // 获取目标数据
        const targetTodo = await Todo.findOne({
            _id: new ObjectId(req.query.todoId as string),
            userId: new ObjectId(userId)
        }).exec();
        const targetEvents = await Event.find({
            todoId: new ObjectId(req.query.todoId as string),
            userId: new ObjectId(userId)
        }).exec();

        // 判断获取结果
        if (!targetTodo) {
            res.json(useErrorResponseData('目标数据不存在，请检查参数'));
            return;
        }

        // 复制待办任务
        const newTodo = await Todo.create({
            userId: new ObjectId(userId),
            projectId: targetTodo.projectId,
            name: targetTodo.name + " 的复制",
            description: targetTodo.description,
            dueDate: targetTodo.dueDate,
            priority: targetTodo.priority,
            state: targetTodo.state,
            tags: targetTodo.tags,
            isFavorited: targetTodo.isFavorited,
            isArchived: targetTodo.isArchived,
            isDeleted: targetTodo.isDeleted
        });

        // 检测复制结果
        if (!newTodo) {
            res.json(useErrorResponseData('复制失败'));
            return;
        }

        // 复制检查事项
        if (targetEvents.length) {
            await Event.insertMany(
                targetEvents.map(event => {
                    return {
                        userId: new ObjectId(userId),
                        todoId: newTodo._id,
                        title: event.title,
                        description: event.description,
                        isDone: event.isDone,
                        doneAt: event.doneAt,
                        isTopped: event.isTopped
                    };
                })
            );
        }

        // 返回复制结果
        res.json(
            useSuccessfulResponseData({
                ...newTodo.toJSON(),
                id: newTodo._id.toString()
            })
        );
    } catch (error) {
        console.log('[api/todo/copyTodo] Error:', error);
        res.json(useResponseData(50001, '复制失败，请稍候重试', null));
    }
};

export { duplicateTodo };
