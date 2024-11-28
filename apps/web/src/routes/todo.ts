import express from 'express';
import {
    createTodo,
    deleteTodo,
    getTodo,
    getTodos,
    updateTodo,
    updateTodos
} from '@nao-todo-server/apis';
// import connectAndRun from '@nao-todo-server/utils/src/connect-and-run';

const exRouter: express.Router = express.Router();

exRouter.post('/todo', async (req, res) => {
    console.log('[/api/todo] Route matched!', `(${req.method})`);
    // await connectAndRun(async () => {
    await createTodo(req, res);
    // });
});

exRouter.delete('/todo', async (req, res) => {
    console.log('[/api/todo] Route matched!', `(${req.method})`);
    // await connectAndRun(async () => {
    await deleteTodo(req, res);
    // });
});

exRouter.put('/todo', async (req, res) => {
    console.log('[/api/todo] Route matched!', `(${req.method})`);
    // await connectAndRun(async () => {
    await updateTodo(req, res);
    // });
});

exRouter.get('/todo', async (req, res) => {
    console.log('[/api/todo] Route matched!', `(${req.method})`);
    // await connectAndRun(async () => {
    await getTodo(req, res);
    // });
});

exRouter.put('/todos', async (req, res) => {
    console.log('[/api/todos] Route matched!', `(${req.method})`);
    // await connectAndRun(async () => {
    await updateTodos(req, res);
    // });
});

exRouter.get('/todos', async (req, res) => {
    console.log('[/api/todos] Route matched!', `(${req.method})`);
    // await connectAndRun(async () => {
    await getTodos(req, res);
    // });
});

export default exRouter;
