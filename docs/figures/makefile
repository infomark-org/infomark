.PHONY: figures clean publish all

FIGURES = $(shell echo *.tex | sed "s/\.tex/\.pdf/g")
IMAGES = $(shell echo *.tex | sed "s/\.tex/\.png/g")
TEX = pdflatex -shell-escape -interaction=nonstopmode -file-line-error

clean:
	rm *.pdf || true
	rm *.png || true

publish:
	cp *.png ../docs/static/images/illustrations/

figures: $(FIGURES)
images: $(IMAGES)

all: clean figures images publish

$(FIGURES):
	$(TEX) $(@:%.pdf=%.tex)

$(IMAGES):
	convert -density 300 -resize 50% $(@:%.png=%.pdf) -quality 100 $(@)