import fileinput

for line in fileinput.input():
    s = line.split(" | ")
    r, t = [], []
    for ix, i in enumerate(s):
        u = [int(j) for j in i.split()]
        for jx in u:
            f = False
            for kx, k in enumerate(t):
                if k[0] != jx:
                    continue
                t[kx][1] += ',{:d}'.format(ix + 1)
                f = True
                break
            if not f:
                t.append([jx, str(ix + 1)])
    for i in sorted(t):
        s = '{}:{};'.format(i[0], i[1])
        r.append(s)
    print(' '.join(r))
