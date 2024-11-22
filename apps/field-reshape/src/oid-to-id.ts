import mongoose from 'mongoose';
import { Event, Project, Todo, Tag } from '@nao-todo-server/models';

const updateOidToIdInProjects = async () => {
    try {
        await mongoose.connect('mongodb://localhost:27017/naotodo');
        const projects = await Project.find({}).exec();
        const updateTasks: Promise<any>[] = [];
        for (const project of projects) {
            const updateTask = Project.findOneAndUpdate(
                { _id: project._id },
                { $set: { userId: project.userId.toString() } },
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

const updateOidToIdInTodos = async () => {
    try {
        await mongoose.connect('mongodb://localhost:27017/naotodo');
        const todos = await Todo.find({}).exec();
        const updateTasks: Promise<any>[] = [];
        for (const todo of todos) {
            const updateTask = Todo.findOneAndUpdate(
                { _id: todo._id },
                {
                    $set: {
                        userId: todo.userId.toString(),
                        projectId: todo.projectId.toString()
                    }
                },
                { new: true }
            ).exec();
            updateTasks.push(updateTask);
        }
        const result = await Promise.all(updateTasks);
        console.log(result);
    } catch (error) {
        console.error('[updateOidToIdInTodos]', error);
    } finally {
        await mongoose.disconnect();
    }
};

const updateOidToIdInEvents = async () => {
    try {
        await mongoose.connect('mongodb://localhost:27017/naotodo');
        const events = await Event.find({}).exec();
        const updateTasks: Promise<any>[] = [];
        for (const event of events) {
            const updateTask = Event.findOneAndUpdate(
                { _id: event._id },
                {
                    $set: {
                        userId: '66d6f5944ae76057cfaa3570',
                        todoId: event.todoId.toString()
                    }
                },
                { new: true }
            ).exec();
            updateTasks.push(updateTask);
        }
        const result = await Promise.all(updateTasks);
        console.log(result);
    } catch (error) {
        console.error('[updateOidToIdInEvents]', error);
    } finally {
        await mongoose.disconnect();
    }
};

const updateOidToIdInTags = async () => {
    try {
        await mongoose.connect('mongodb://localhost:27017/naotodo');
        const tags = await Tag.find({}).exec();
        const updateTasks: Promise<any>[] = [];
        for (const tag of tags) {
            const updateTask = Tag.findOneAndUpdate(
                { _id: tag._id },
                { $set: { userId: tag.userId.toString() } },
                { new: true }
            ).exec();
            updateTasks.push(updateTask);
        }
        const result = await Promise.all(updateTasks);
        console.log(result);
    } catch (error) {
        console.error('[updateOidToIdInEvents]', error);
    } finally {
        await mongoose.disconnect();
    }
};

export {
    updateOidToIdInProjects,
    updateOidToIdInTodos,
    updateOidToIdInEvents,
    updateOidToIdInTags
};
