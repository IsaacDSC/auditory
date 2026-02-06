MODULE_PATH=$(grep '^module ' go.mod | awk '{print $2}')

echo "digraph G {" > imports.dot
echo "  rankdir=LR;" >> imports.dot
echo "  node [shape=box, fontname=\"Arial\", fontsize=10];" >> imports.dot

awk -v module="$MODULE_PATH" '
{
  pkg = $1
  for (i=2; i<=NF; i++) {
    imp = $i

    # se quiser incluir TUDO (interno + externo), comente o if abaixo
    # e deixe só o print
    # if (index(imp, module) != 1) {
    #   # pula imports externos se quiser só internos
    #   next
    # }

    from = pkg
    to   = imp

    # remove prefixo do módulo para ficar mais legível
    sub(module"/", "", from)
    sub(module"/", "", to)

    print "  \"" from "\" -> \"" to "\";"
  }
}
' imports.txt >> imports.dot

echo "}" >> imports.dot
