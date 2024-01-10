
typedef struct GB GB;
typedef unsigned (InputGetter)(void *);

extern int gambatte_revision();
extern GB * gambatte_create();
extern void gambatte_destroy(GB *g);
extern int gambatte_load(GB *g, char const *romfile, unsigned flags);
extern int gambatte_loadbios(GB *g, char const *biosfile, unsigned size, unsigned crc);
extern int gambatte_runfor(GB *g, unsigned *videoBuf, int pitch, unsigned *audioBuf, unsigned *samples);
extern void gambatte_reset(GB *g, unsigned samplesToStall);
extern void gambatte_setinputgetter(GB *g, InputGetter *getInput, void *p);
extern unsigned gambatte_savestate(GB *g, unsigned const *videoBuf, int pitch, char *stateBuf);
extern int gambatte_loadstate(GB *g, char const *stateBuf, unsigned size);
extern unsigned gambatte_timenow(GB *g);

extern int input_callback(void *p);