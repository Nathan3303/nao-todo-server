import express from 'express';
import {
    createEvent,
    deleteEvent,
    getEvent,
    getEvents,
    updateEvent
} from '@nao-todo-server/apis';
// import connectAndRun from '@nao-todo-server/utils/src/connect-and-run';

const exRouter: express.Router = express.Router();

exRouter.post('/event', async (req, res) => {
    console.log('[/api/event] Route matched!', `(${req.method})`);
    // await connectAndRun(async () => {
    await createEvent(req, res);
    // });
});

exRouter.delete('/event', async (req, res) => {
    console.log('[/api/event] Route matched!', `(${req.method})`);
    // await connectAndRun(async () => {
    await deleteEvent(req, res);
    // });
});

exRouter.put('/event', async (req, res) => {
    console.log('[/api/event] Route matched!', `(${req.method})`);
    // await connectAndRun(async () => {
    await updateEvent(req, res);
    // });
});

exRouter.get('/event', async (req, res) => {
    console.log('[/api/event] Route matched!', `(${req.method})`);
    // await connectAndRun(async () => {
    await getEvent(req, res);
    // });
});

exRouter.get('/events', async (req, res) => {
    console.log('[/api/events] Route matched!', `(${req.method})`);
    // await connectAndRun(async () => {
    await getEvents(req, res);
    // });
});

export default exRouter;
