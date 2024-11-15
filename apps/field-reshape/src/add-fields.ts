import mongoose from 'mongoose';
import { Event, Project, Todo, Tag } from '@nao-todo-server/models';
import { ObjectId } from '@nao-todo-server/utils';

const addFieldsInTodos = async () => {
    try {
        await mongoose.connect('mongodb://localhost:27017/naotodo');
        const todos = await Todo.find({}).exec();
        const updateTasks: Promise<any>[] = [];
        for (const todo of todos) {
            const updateTask = Todo.findOneAndUpdate(
                { _id: todo._id },
                { $set: { isFavorited: false } },
                { new: true }
            ).exec();
            updateTasks.push(updateTask);
        }
        const result = await Promise.all(updateTasks);
        console.log(result);
    } catch (error) {
        console.error('[updateOidToIdInProjects]', error);
    } finally {
        await mongoose.disconnect();
    }
};

export { addFieldsInTodos };
