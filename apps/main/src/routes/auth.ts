import express from 'express';
import connectAndRun from '@nao-todo-server/utils/src/connect-and-run';
import { signin, signup, checkin, signout } from '@nao-todo-server/apis';
// import type { Request, Response } from 'express';

const exRouter: express.Router = express.Router();

exRouter.post('/signin', async (req, res) => {
    console.log('[/api/signin] Route matched!');
    await connectAndRun(async () => {
        await signin(req, res);
    });
});

exRouter.post('/signup', async (req, res) => {
    console.log('[/api/signup] Route matched!');
    await connectAndRun(async () => {
        await signup(req, res);
    });
});

exRouter.get('/checkin', async (req, res) => {
    console.log('[/api/checkin] Route matched!');
    await connectAndRun(async () => {
        await checkin(req, res);
    });
});

exRouter.delete('/signout', async (req, res) => {
    console.log('[/api/signout] Route matched!');
    await connectAndRun(async () => {
        await signout(req, res);
    });
});

export default exRouter;
